package servicehttp

import (
	"context"
	"log/slog"
	"net/http"
	"net/url"
	"slices"
	"strings"
	"sync"
	"time"

	pb "github.com/Flikest/PingVi_backend/gen/go/user_info"
	"github.com/Flikest/PingVi_backend/internal/delivery/dto"
	"github.com/Flikest/PingVi_backend/internal/repository"
	mimetype "github.com/Flikest/PingVi_backend/pkg/mime_type"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024 * 2,
	WriteBufferSize: 1024 * 2,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

type Message struct {
	Operation   string      `json:"operation"`
	Message     dto.Message `json:"message"`
	CommunityID uuid.UUID   `json:"community_id"`
}

type Client struct {
	ID     string
	Conn   websocket.Conn
	Send   chan Message
	ChatID string
}

type Hub struct {
	Online              map[string]*Client
	RepositoryMessenger *repository.RepositoryMessenger
	Log                 *slog.Logger

	mu sync.RWMutex
}

func NewHub(h *Hub) *Hub {
	return &Hub{
		Online:              h.Online,
		RepositoryMessenger: h.RepositoryMessenger,
		Log:                 h.Log,
	}
}

func (h *Hub) onConnect(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.Online[client.ID] = client
}

func (h *Hub) onDisconect(userID string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	_, exist := h.Online[userID]

	if exist {
		delete(h.Online, userID)
	}
}
func (h *Hub) onSwitchChat(userID, swichedChatID uuid.UUID) {
	h.mu.Lock()
	defer h.mu.Unlock()

	value, exist := h.Online[userID.String()]

	if exist {
		if value.ChatID != swichedChatID.String() {
			h.Online[userID.String()].ChatID = swichedChatID.String()
		}
	}
}

func (h *Hub) onRead(r dto.ReadMessage) {
	ctx := context.Background()

	r.At = time.Now()

	h.RepositoryMessenger.ReadMessage(ctx, r)
}

func (h *Hub) onSendMessage(msg dto.AddMessage) {
	ctx := context.Background()

	messageID, err := uuid.NewV7()
	if err != nil {
		h.Log.Error("error with generating message id: ", "error", err)
		return
	}

	now := time.Now()

	message := dto.Message{
		ID:          messageID,
		ChatID:      msg.ChatID,
		SenderID:    msg.SenderID,
		Message:     msg.Message,
		MessageType: msg.MessageType,
		Attachments: msg.Attachments,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := h.RepositoryMessenger.InsertMessage(ctx, message); err != nil {
		h.Log.Error("error with inserting message: ", "error", err)
		return
	}

	go h.broadcast(ctx, Message{
		Operation: "add",
		Message:   message,
	})
}

func (h *Hub) onUpdateMesage(msg dto.UpdateMessage) {
	ctx := context.Background()

	now := time.Now()

	message := dto.Message{
		ID:          msg.ID,
		ChatID:      msg.ChatID,
		SenderID:    msg.UserID,
		Message:     msg.Message,
		MessageType: msg.MessageType,
		IsEdited:    true,
		ReplyToID:   msg.ReplyToID,
		Attachments: msg.Attachments,
		Reactions:   msg.Reactions,
		CreatedAt:   msg.CreatedAt,
		UpdatedAt:   now,
	}

	if err := h.RepositoryMessenger.UpdateMessage(ctx, message); err != nil {
		h.Log.Error("error with updating message: ", "error", err)
		return
	}

	go h.broadcast(ctx, Message{
		Operation: "update",
		Message:   message,
	})
}

func (h *Hub) onDeleteMessage(msg dto.DeleteMessage) {
	ctx := context.Background()

	if err := h.RepositoryMessenger.DeleteMessage(ctx, msg.ID, msg.ChatID, msg.SenderID); err != nil {
		h.Log.Error("error with deleting message: ", "error", err)
		return
	}

	go h.broadcast(ctx, Message{
		Operation: "delete",
		Message: dto.Message{
			ID: msg.ID,
		},
	})
}

func (h *Hub) broadcast(ctx context.Context, message Message) {
	participants, err := h.RepositoryMessenger.SelectAllMembersGroup(ctx, message.Message.ChatID)
	if err != nil {
		h.Log.Error("failed to retrieve all messages from the chat: ", "error", err)
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, j := range participants {
		client, exists := h.Online[j.UserID.String()]
		if !exists {
			continue // TODO заменить на отправку push уведомления
		}

		client.Send <- message
	}
}

func (s *ServiceMessenger) ReadMessageFromClient(client *Client) {
	defer func() {
		s.Hub.onDisconect(client.ID)
		client.Conn.Close()
	}()

	for {
		var msg Message

		// TODO получить user id по session id

		if err := client.Conn.ReadJSON(&msg); err != nil {
			s.Hub.Log.Error("error with reading message: ", "error", err)
			break
		}

		switch msg.Operation {
		case "switch_community":
			s.Hub.onSwitchChat(msg.Message.SenderID, msg.CommunityID)
		case "read":
			s.Hub.onRead(dto.ReadMessage{
				ChatID:    msg.Message.ChatID,
				UserID:    msg.Message.SenderID,
				MessageID: msg.Message.ID,
				At:        time.Now(),
			})
		case "send":
			s.Hub.onSendMessage(dto.AddMessage{
				ChatID:      msg.Message.ChatID,
				SenderID:    msg.Message.SenderID,
				Message:     msg.Message.Message,
				MessageType: msg.Message.MessageType,
				Attachments: msg.Message.Attachments,
			})
		case "update":
			s.Hub.onUpdateMesage(dto.UpdateMessage{
				ID:          msg.Message.ID,
				ChatID:      msg.Message.ChatID,
				SenderID:    msg.Message.SenderID,
				Message:     msg.Message.Message,
				MessageType: msg.Message.MessageType,
				ReplyToID:   msg.Message.ReplyToID,
				Attachments: msg.Message.Attachments,
				Reactions:   msg.Message.Reactions,
				CreatedAt:   msg.Message.CreatedAt,
			})
		case "delete":
			s.Hub.onDeleteMessage(dto.DeleteMessage{
				ID:       msg.Message.ID,
				ChatID:   msg.Message.ChatID,
				SenderID: msg.Message.SenderID,
			})
		}
	}
}

func (s *ServiceMessenger) writeMessageToClient(client *Client) {
	defer func() {
		close(client.Send)
		client.Conn.Close()
	}()

	for message := range client.Send {
		if err := client.Conn.WriteJSON(message); err != nil {
			s.Log.Error("error writing to client: ", "error", err)
			break
		}
	}
}

// Handshake godoc
//
//	@Summary		WebSocket connection for real-time messaging
//	@Description	Establishes a WebSocket connection for real-time messaging.
//	@Description	- Requires session_id for authentication
//	@Description	- Supports operations: switch_community, read, send, update, delete
//	@Description	- Messages are broadcasted to all members of the chat
//	@Tags			messenger-websocket
//	@Accept			json
//	@Produce		json
//	@Param			session_id	path		string					true	"Session ID for authentication"	example("abc123def456")
//	@Success		101			{object}	websocket.Conn			"WebSocket connection upgraded"
//	@Failure		400			{object}	map[string]interface{}	"Invalid session ID"		example({"error":"invalid session id"})
//	@Failure		415			{object}	map[string]interface{}	"Invalid content type"		example({"error":"Content-Type must be application/json"})
//	@Failure		500			{object}	map[string]interface{}	"WebSocket upgrade failed"	example({"error":"WebSocket upgrade failed"})
//	@Router			/ws/{session_id} [get]
func (s *ServiceMessenger) Handshake(ctx *gin.Context) {
	defer ctx.Request.Body.Close()

	if ctx.GetHeader("Content-Type") != "application/json" {
		s.Log.Error("invalid content type")
		ctx.Status(http.StatusUnsupportedMediaType)
		ctx.JSON(http.StatusUnsupportedMediaType, map[string]string{
			"error": "Content-Type must be application/json",
		})
		return

	}

	payload, err := s.Client.GetUserIDBySessionID(ctx.Request.Context(), &pb.GetUserIDBySessionIDRequest{
		SessionId: ctx.Param("session_id"),
	})

	conn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		s.Log.Error("error with handshake to client: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, "WebSocket upgrade failed")
		return
	}

	client := Client{
		ID:     payload.UserId,
		Conn:   *conn,
		Send:   make(chan Message, 256),
		ChatID: "main",
	}

	s.Hub.onConnect(&client)

	go s.ReadMessageFromClient(&client)
	go s.writeMessageToClient(&client)
}

// LinkPreview godoc
//
//	@Summary		Get link preview
//	@Description	Fetches a preview for a given link. Supports internal PingVi links and external URLs.
//	@Description	- Internal links: returns type and message
//	@Description	- External links: fetches content type and returns preview
//	@Tags			messenger-websocket
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			Authorization	header		string					true					"Bearer JWT token"	example("Bearer eyJhbGciOiJIUzI1NiIs...")
//	@Param			link			path		string					true					"URL to preview"	example("https://example.com/image.jpg")
//	@Success		200				{object}	dto.LinkPreviewResponse	"Link preview data"		example({"type":"image","message":"https://s3.amazonaws.com/public/emoji/standart/1F600.json"})
//	@Failure		400				{object}	map[string]interface{}	"Invalid URL"			example({"error":"invalid url format"})
//	@Failure		500				{object}	map[string]interface{}	"Internal server error"	example({"error":"failed to fetch preview"})
//	@Router			/ws/preview/{link} [get]
func (s *ServiceMessenger) LinkPreview(ctx *gin.Context) {
	link := ctx.Param("link")

	u, err := url.Parse(link)
	if err != nil {
		s.Log.Error("error with parsing link: ", "error", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid url format"})
		return
	}

	internalServices := []string{"groups", "channels", "s3"}

	// TODO заменить на настоящий host
	if u.Host == "pingvi.com" {
		path := strings.Split(strings.TrimPrefix(u.Path, "/"), "/")
		if len(path) >= 2 && slices.Contains(internalServices, path[0]) {
			ctx.JSON(http.StatusOK, dto.LinkPreviewResponse{
				Type:    path[1],
				Message: link,
			})
			return
		}
	}

	request, err := http.NewRequestWithContext(ctx.Request.Context(), http.MethodHead, link, nil)
	if err != nil {
		s.Log.Error("request error:", "error", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	client := http.Client{
		Timeout: 5 * time.Second,
	}

	response, err := client.Do(request)
	if err != nil {
		s.Log.Error("error while requesting URL: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer response.Body.Close()

	contentType := strings.Split(response.Header.Get("Content-Type"), ";")[0]
	contentType = strings.TrimSpace(strings.ToLower(contentType))

	previewType := mimetype.GetMimeType(contentType)

	if previewType == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "A file with this extension is not supported by PingVi"})
		return
	}

	ctx.JSON(http.StatusOK, dto.LinkPreviewResponse{
		Type:    previewType,
		Message: link,
	})
}

// GetAllMessageFromChat godoc
//
//	@Summary		Get all messages from a chat
//	@Description	Retrieves all messages from a specific chat (channel, group, or personal chat).
//	@Tags			messenger-websocket
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			Authorization	header		string					true	"Bearer JWT token"	example("Bearer eyJhbGciOiJIUzI1NiIs...")
//	@Param			chat_id			path		string					true	"Chat ID"			example("123e4567-e89b-12d3-a456-426614174000")
//	@Success		200				{array}		dto.Message				"List of messages"
//	@Failure		400				{object}	map[string]interface{}	"Invalid chat ID"		example({"error":"invalid chat id"})
//	@Failure		500				{object}	map[string]interface{}	"Internal server error"	example({"error":"failed to get messages"})
//	@Router			/ws/messages/{chat_id} [get]
func (s *ServiceMessenger) GetAllMessageFromChat(ctx *gin.Context) {
	chatID, err := uuid.Parse(ctx.Param("chat_id"))
	if err != nil {
		s.Log.Error("error with parsing chat id uuid: ", "error", err)
		ctx.JSON(http.StatusBadRequest, err)
		return
	}

	messages, err := s.Repository.SelectMessagesByChatID(ctx.Request.Context(), chatID)
	if err != nil {
		s.Log.Error("error with selecting all messages from chat: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, messages)
}

// ClearMesagesFromChat godoc
//
//	@Summary		Clear all messages from a chat
//	@Description	Clears all messages from a specific chat. Requires appropriate permissions.
//	@Tags			messenger-websocket
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			Authorization	header		string					true					"Bearer JWT token"	example("Bearer eyJhbGciOiJIUzI1NiIs...")
//	@Param			request			body		dto.ClearMessage		true					"Chat ID"			example({"chat_id":"123e4567-e89b-12d3-a456-426614174000"})
//	@Success		200				{object}	map[string]interface{}	"Chat ID"				example({"chat_id":"123e4567-e89b-12d3-a456-426614174000"})
//	@Failure		400				{object}	map[string]interface{}	"Invalid request"		example({"error":"invalid body"})
//	@Failure		500				{object}	map[string]interface{}	"Internal server error"	example({"error":"failed to clear messages"})
//	@Router			/ws/messages/clear [delete]
func (s *ServiceMessenger) ClearMesagesFromChat(ctx *gin.Context) {
	var body dto.ClearMessage
	if err := ctx.BindJSON(&body); err != nil {
		s.Log.Error("invalid body", "error", err)
		ctx.JSON(http.StatusBadRequest, err)
		return
	}

	chatID, err := s.ClearMessagesFromCommunity(ctx, body)
	if err != nil {
		s.Log.Error("error with clearing message from chat: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"chat_id": chatID})
}
