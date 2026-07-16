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

	"github.com/Flikest/PingVi_backend/gen/go/user_info"
	"github.com/Flikest/PingVi_backend/internal/delivery/dto"
	"github.com/Flikest/PingVi_backend/internal/repository"
	mimetype "github.com/Flikest/PingVi_backend/pkg/mime_type"
	"github.com/bwmarrin/snowflake"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024 * 2,
	WriteBufferSize: 1024 * 2,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

type Message struct {
	Operation string      `json:"operation"`
	Message   dto.Message `json:"message"`
	Reaction  string      `json:"reaction"`
	Status    string      `json:"status"`
}

type Client struct {
	ID     string
	Conn   *websocket.Conn
	Send   chan Message
	ChatID string

	mu       sync.Mutex
	isClosed bool
}

func (c *Client) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.isClosed {
		return
	}
	c.isClosed = true
	c.Conn.Close()
	close(c.Send)
}

func (c *Client) Push(msg Message) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.isClosed {
		return false
	}

	select {
	case c.Send <- msg:
		return true
	default:
		c.Conn.Close()
		return false
	}
}

type Hub struct {
	Clients             map[string]*Client
	Redis               *redis.Client
	RepositoryMessenger *repository.RepositoryMessenger
	Log                 *slog.Logger
	Node                *snowflake.Node

	mu sync.RWMutex
}

func NewHub(h *Hub) *Hub {
	return &Hub{
		Clients:             h.Clients,
		Redis:               h.Redis,
		RepositoryMessenger: h.RepositoryMessenger,
		Log:                 h.Log,
		Node:                h.Node,
	}
}

func (h *Hub) onConnect(ctx context.Context, client *Client, subscribeUsers []uuid.UUID) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.Clients[client.ID] = client

	userID, err := uuid.Parse(client.ID)
	if err == nil && h.Redis != nil {
		h.setUserOnline(ctx, userID)

		for _, subUserID := range subscribeUsers {
			h.addUserSubscription(ctx, userID, subUserID)
		}

		go h.sendInitialStatuses(ctx, userID, client)
	}
}

func (h *Hub) onDisconect(ctx context.Context, userID string) {
	h.mu.Lock()
	client, exist := h.Clients[userID]
	if exist {
		delete(h.Clients, userID)
	}
	h.mu.Unlock()

	if h.Redis != nil && exist {
		userUUID, err := uuid.Parse(userID)
		if err == nil {
			h.setUserOffline(ctx, userUUID)
		}
	}

	if exist {
		client.Close()
	}
}

func (h *Hub) onSwitchChat(ctx context.Context, userID, swichedChatID uuid.UUID) {
	h.mu.Lock()
	defer h.mu.Unlock()

	client, exist := h.Clients[userID.String()]
	if exist {
		if client.ChatID != swichedChatID.String() {
			client.ChatID = swichedChatID.String()
		}
	}

	if h.Redis != nil && exist {
		h.setUserActiveChat(ctx, userID, swichedChatID)

		go h.notifySubscribers(ctx, userID, "active_chat", swichedChatID)
	}
}

func (h *Hub) onRead(ctx context.Context, r dto.ReadMessage) {
	r.At = time.Now()
	h.RepositoryMessenger.ReadMessage(ctx, r)
}

func (h *Hub) onSendMessage(ctx context.Context, msg dto.AddMessage) {
	messageID := h.Node.Generate()

	now := time.Now()
	message := dto.Message{
		ID:        messageID,
		ChatID:    msg.ChatID,
		SenderID:  msg.SenderID,
		Message:   msg.Message,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := h.RepositoryMessenger.InsertMessage(ctx, message); err != nil {
		h.Log.Error("error with inserting message", "error", err)
		return
	}

	go h.broadcast(ctx, Message{
		Operation: "add",
		Message:   message,
	})
}

func (h *Hub) onUpdateMesage(ctx context.Context, msg dto.UpdateMessage) {
	now := time.Now()
	message := dto.Message{
		ID:        msg.ID,
		ChatID:    msg.ChatID,
		SenderID:  msg.UserID,
		Message:   msg.Message,
		ReplyToID: msg.ReplyToID,
		CreatedAt: msg.CreatedAt,
		UpdatedAt: now,
	}

	if err := h.RepositoryMessenger.UpdateMessage(ctx, message); err != nil {
		h.Log.Error("error with updating message", "error", err)
		return
	}

	go h.broadcast(ctx, Message{
		Operation: "update",
		Message:   message,
	})
}

func (h *Hub) onSendReaction(ctx context.Context, reaction dto.Reaction) {
	if err := h.RepositoryMessenger.InsertReaction(ctx, reaction); err != nil {
		h.Log.Error("error with set up reaction", "error", err)
		return
	}

	go h.broadcast(ctx, Message{
		Operation: "send_reaction",
		Message: dto.Message{
			ID:        reaction.MessageID,
			ChatID:    reaction.ChatID,
			SenderID:  reaction.UserID,
			CreatedAt: reaction.SendedAt,
		},
		Reaction: reaction.Reaction,
	})
}

func (h *Hub) onDeleteReaction(ctx context.Context, chatID uuid.UUID, messageID snowflake.ID, userID uuid.UUID) {
	if err := h.RepositoryMessenger.DeleteReaction(ctx, messageID, chatID, userID); err != nil {
		h.Log.Error("error with delete reaction", "error", err)
		return
	}

	go h.broadcast(ctx, Message{
		Operation: "delete_reaction",
		Message: dto.Message{
			ID:       messageID,
			ChatID:   chatID,
			SenderID: userID,
		},
	})
}

func (h *Hub) onDeleteMessage(ctx context.Context, msg dto.DeleteMessage) {
	if err := h.RepositoryMessenger.DeleteMessage(ctx, msg.ID, msg.ChatID, msg.SenderID); err != nil {
		h.Log.Error("error with deleting message", "error", err)
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
		h.Log.Error("failed to retrieve all messages from the chat", "error", err)
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, j := range participants {
		client, exists := h.Clients[j.UserID.String()]
		if !exists {
			continue
		}

		message.Message.IsMy = message.Message.SenderID.String() == j.UserID.String()

		client.Push(message)
	}
}

func (h *Hub) sendInitialStatuses(ctx context.Context, userID uuid.UUID, client *Client) {
	if h.Redis == nil {
		return
	}

	statuses, err := h.getSubscribedUsersStatus(ctx, userID)
	if err != nil {
		h.Log.Error("failed to get subscribed users status", "error", err)
		return
	}

	for subUserID, _ := range statuses {
		statusMessage := Message{
			Operation: "user_status",
			Message: dto.Message{
				SenderID: subUserID,
			},
		}

		client.Push(statusMessage)
	}
}

func (h *Hub) notifySubscribers(ctx context.Context, userID uuid.UUID, statusType string, chatID uuid.UUID) {
	if h.Redis == nil {
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, client := range h.Clients {
		clientUserID, err := uuid.Parse(client.ID)
		if err != nil {
			continue
		}

		subs, err := h.getUserSubscriptions(ctx, clientUserID)
		if err != nil {
			continue
		}

		for _, subID := range subs {
			if subID == userID {
				statusMessage := Message{
					Operation: "user_status_update",
					Message: dto.Message{
						SenderID: userID,
						ChatID:   chatID,
					},
				}
				if statusType == "typing" {
					statusMessage.Status = "typing_start"
				} else if statusType == "active_chat" {
					statusMessage.Status = "active_chat_changed"
				}

				client.Push(statusMessage)
				break
			}
		}
	}
}

func (s *ServiceMessenger) ReadMessageFromClient(ctx context.Context, client *Client) {
	readCtx, cancel := context.WithCancel(ctx)
	defer func() {
		cancel()
		s.Hub.onDisconect(ctx, client.ID)
	}()

	go func() {
		<-readCtx.Done()
		client.Conn.Close()
	}()

	for {
		var msg Message
		if err := client.Conn.ReadJSON(&msg); err != nil {
			if readCtx.Err() != nil {
				return
			}
			s.Hub.Log.Error("error with reading message", "error", err)
			return
		}

		switch msg.Operation {
		case "switch_community":
			s.Hub.onSwitchChat(readCtx, msg.Message.SenderID, msg.Message.ChatID)
		case "read":
			s.Hub.onRead(readCtx, dto.ReadMessage{
				ChatID:    msg.Message.ChatID,
				UserID:    msg.Message.SenderID,
				MessageID: msg.Message.ID,
				At:        time.Now(),
			})
		case "send":
			s.Hub.onSendMessage(readCtx, dto.AddMessage{
				ChatID:   msg.Message.ChatID,
				SenderID: msg.Message.SenderID,
				Message:  msg.Message.Message,
			})
		case "update":
			s.Hub.onUpdateMesage(readCtx, dto.UpdateMessage{
				ID:        msg.Message.ID,
				ChatID:    msg.Message.ChatID,
				SenderID:  msg.Message.SenderID,
				Message:   msg.Message.Message,
				ReplyToID: msg.Message.ReplyToID,
				CreatedAt: msg.Message.CreatedAt,
			})
		case "delete":
			s.Hub.onDeleteMessage(readCtx, dto.DeleteMessage{
				ID:       msg.Message.ID,
				ChatID:   msg.Message.ChatID,
				SenderID: msg.Message.SenderID,
			})
		case "send_reaction":
			s.Hub.onSendReaction(readCtx, dto.Reaction{
				ChatID:    msg.Message.ChatID,
				MessageID: msg.Message.ID,
				UserID:    msg.Message.SenderID,
				Reaction:  msg.Reaction,
				SendedAt:  time.Now(),
			})
		case "delete_reaction":
			s.Hub.onDeleteReaction(readCtx, msg.Message.ChatID, msg.Message.ID, msg.Message.SenderID)
		case "typing_start":
			if s.Hub.Redis != nil {
				s.Hub.setUserTyping(readCtx, msg.Message.SenderID, msg.Message.ChatID)
				go s.Hub.notifySubscribers(readCtx, msg.Message.SenderID, "typing", msg.Message.ChatID)
			}
		case "typing_stop":
			if s.Hub.Redis != nil {
				s.Hub.clearUserTyping(readCtx, msg.Message.SenderID)
			}
		}
	}
}

func (s *ServiceMessenger) writeMessageToClient(ctx context.Context, client *Client) {
	for {
		select {
		case <-ctx.Done():
			return
		case message, ok := <-client.Send:
			if !ok {
				return
			}

			client.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := client.Conn.WriteJSON(message); err != nil {
				s.Log.Error("error writing to client", "error", err)
				return
			}
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
	sessionID := ctx.Param("session_id")
	if sessionID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid session id"})
		return
	}

	response, err := s.Client.GetUserIDBySessionID(ctx.Request.Context(), &user_info.GetUserIDBySessionIDRequest{
		SessionId: sessionID,
	})
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "error with getting user id by session id"})
		return
	}

	userID, err := uuid.Parse(response.GetUserId())
	if err != nil {
		s.Log.Error("error with parsing user id uuid: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error with parsing user id"})
	}

	wsConn, err := upgrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		s.Log.Error("WebSocket upgrade failed", "error", err)
		return
	}

	subscribers, err := s.GetSubscribers(ctx.Request.Context(), userID)

	client := &Client{
		ID:   sessionID,
		Conn: wsConn,
		Send: make(chan Message, 256),
	}

	reqCtx := ctx.Request.Context()

	s.Hub.onConnect(reqCtx, client, subscribers)

	go s.writeMessageToClient(reqCtx, client)

	s.ReadMessageFromClient(reqCtx, client)
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

	chatID, err := s.ClearMessagesFromCommunity(ctx.Request.Context(), body)
	if err != nil {
		s.Log.Error("error with clearing message from chat: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"chat_id": chatID})
}
