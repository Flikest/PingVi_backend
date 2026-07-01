package servicehttp

import (
	"fmt"
	"net/http"
	"os"

	pb "github.com/Flikest/PingVi_backend/gen/go/user_info"
	"github.com/Flikest/PingVi_backend/internal/delivery/dto"
	"github.com/Flikest/PingVi_backend/pkg/tokens"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// CreateTopic godoc
// @Summary      Create a new topic in a channel
// @Description  Creates a new topic (text or voice) in a channel. Requires appropriate permissions.
// @Description  - Topic types: "text" or "voice"
// @Description  - System message will be sent to the channel
// @Tags         messenger-topics
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        Authorization  header  string  true  "Bearer JWT token"  example("Bearer eyJhbGciOiJIUzI1NiIs...")
// @Param        request  body  dto.CreateTopic  true  "Topic creation data"  example({"channel_id":"123e4567-e89b-12d3-a456-426614174000","name":"General Chat","topic_type":"text"})
// @Success      201  {object}  dto.Topic  "Created topic"
// @Failure      400  {object}  map[string]interface{}  "Invalid request"  example({"error":"invalid body"})
// @Failure      500  {object}  map[string]interface{}  "Internal server error"  example({"error":"failed to create topic"})
// @Router       /topics [post]
func (s *ServiceMessenger) CreateTopic(ctx *gin.Context) {
	payload, err := tokens.Verify(ctx.Request.Header.Get("Authorization"), []byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		s.Log.Error("error with verify jwt user token: ", "error", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, err := uuid.Parse(payload.ID.String())
	if err != nil {
		s.Log.Error("error with parsing jwt token from header: ", "error", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var body dto.CreateTopic
	if err := ctx.BindJSON(&body); err != nil {
		s.Log.Error("invalid body: ", "error", err)
		ctx.JSON(http.StatusBadRequest, err)
		return
	}

	body.UserID = userID

	topic, err := s.Repository.CreateTopic(ctx.Request.Context(), body)
	if err != nil {
		s.Log.Error("error with creating topic: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}

	response, err := s.Client.GetUserNameByID(ctx.Request.Context(), &pb.GetUserNameByIdRequest{
		UserId: userID.String(),
	})
	if err != nil {
		s.Log.Error("error with getting user name: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var systemMessage string

	if body.TopicType == "text" {
		systemMessage = fmt.Sprintf("User %s created a text topic named %s", response.GetName(), body.Name)
	} else {
		systemMessage = fmt.Sprintf("User %s created a voice topic named %s", response.GetName(), body.Name)
	}

	s.Hub.onSendMessage(dto.AddMessage{
		ChatID:      body.ChannelId,
		SenderID:    uuid.Nil,
		Message:     systemMessage,
		MessageType: "system",
	})

	ctx.JSON(http.StatusCreated, topic)
}

// GetAllTopics godoc
// @Summary      Get all topics in a channel
// @Description  Retrieves all topics from a specific channel. User must be a member of the channel.
// @Tags         messenger-topics
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        Authorization  header  string  true  "Bearer JWT token"  example("Bearer eyJhbGciOiJIUzI1NiIs...")
// @Param        channel_id  path  string  true  "Channel ID"  example("123e4567-e89b-12d3-a456-426614174000")
// @Success      200  {array}  dto.Topic  "List of topics"
// @Failure      400  {object}  map[string]interface{}  "Invalid channel ID"  example({"error":"invalid channel id"})
// @Failure      403  {object}  map[string]interface{}  "Not a member"  example({"error":"the user does not belong to this group"})
// @Failure      500  {object}  map[string]interface{}  "Internal server error"  example({"error":"failed to get topics"})
// @Router       /topics/{channel_id} [get]
func (s *ServiceMessenger) GetAllTopics(ctx *gin.Context) {
	payload, err := tokens.Verify(ctx.Request.Header.Get("Authorization"), []byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		s.Log.Error("error with verify jwt user token: ", "error", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, err := uuid.Parse(payload.ID.String())
	if err != nil {
		s.Log.Error("error with parsing jwt token from header: ", "error", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	channelID, err := uuid.Parse(ctx.Param("channel_id"))
	if err != nil {
		s.Log.Error("error with parsing user id uuid: ", "error", err)
		ctx.JSON(http.StatusBadRequest, err)
		return
	}

	isBelong, err := s.Repository.IsBelongChannel(ctx.Request.Context(), channelID, userID)

	if !isBelong {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "the user does not belong to this group"})
	}

	topics, err := s.Repository.SelectTopicsFromChannel(ctx.Request.Context(), channelID)
	if err != nil {
		s.Log.Error("error with getting topics from channel: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, topics)
}

// UpdateTopic godoc
// @Summary      Update a topic
// @Description  Updates an existing topic in a channel. Requires appropriate permissions.
// @Description  - Can update: name, topic type
// @Tags         messenger-topics
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        Authorization  header  string  true  "Bearer JWT token"  example("Bearer eyJhbGciOiJIUzI1NiIs...")
// @Param        request  body  dto.UpdateTopic  true  "Topic update data"  example({"id":"123e4567-e89b-12d3-a456-426614174000","channel_id":"987fcdeb-51d2-12d3-a456-426614174000","name":"New Topic Name"})
// @Success      200  {object}  dto.Topic  "Updated topic"
// @Failure      400  {object}  map[string]interface{}  "Invalid request"  example({"error":"invalid body"})
// @Failure      500  {object}  map[string]interface{}  "Internal server error"  example({"error":"failed to update topic"})
// @Router       /topics [put]
func (s *ServiceMessenger) UpdateTopic(ctx *gin.Context) {
	payload, err := tokens.Verify(ctx.Request.Header.Get("Authorization"), []byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		s.Log.Error("error with verify jwt user token: ", "error", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, err := uuid.Parse(payload.ID.String())
	if err != nil {
		s.Log.Error("error with parsing jwt token from header: ", "error", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var body dto.UpdateTopic
	if err := ctx.BindJSON(&body); err != nil {
		s.Log.Error("invalid body: ", "error", err)
		ctx.JSON(http.StatusBadRequest, err)
		return
	}

	body.UserID = userID

	topic, err := s.Repository.UpdateTopic(ctx.Request.Context(), body)
	if err != nil {
		s.Log.Error("error with updating topic: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, topic)
}

// DeleteTopic godoc
// @Summary      Delete a topic
// @Description  Permanently deletes a topic from a channel. Requires appropriate permissions.
// @Description  - System message will be sent to the channel
// @Tags         messenger-topics
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        Authorization  header  string  true  "Bearer JWT token"  example("Bearer eyJhbGciOiJIUzI1NiIs...")
// @Param        request  body  dto.DeleteTopic  true  "Topic deletion data"  example({"id":"123e4567-e89b-12d3-a456-426614174000","channel_id":"987fcdeb-51d2-12d3-a456-426614174000"})
// @Success      200  {object}  map[string]interface{}  "Topic ID"  example({"topic_id":"123e4567-e89b-12d3-a456-426614174000"})
// @Failure      400  {object}  map[string]interface{}  "Invalid request"  example({"error":"invalid body"})
// @Failure      500  {object}  map[string]interface{}  "Internal server error"  example({"error":"failed to delete topic"})
// @Router       /topics [delete]
func (s *ServiceMessenger) DeleteTopic(ctx *gin.Context) {
	payload, err := tokens.Verify(ctx.Request.Header.Get("Authorization"), []byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		s.Log.Error("error with verify jwt user token: ", "error", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, err := uuid.Parse(payload.ID.String())
	if err != nil {
		s.Log.Error("error with parsing jwt token from header: ", "error", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var body dto.DeleteTopic
	if err := ctx.BindJSON(&body); err != nil {
		s.Log.Error("invalid body: ", "error", err)
		ctx.JSON(http.StatusBadRequest, err)
		return
	}

	body.UserId = userID

	topicID, err := s.Repository.DeleteTopic(ctx.Request.Context(), body)
	if err != nil {
		s.Log.Error("error with updating topic: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}

	response, err := s.Client.GetUserNameByID(ctx.Request.Context(), &pb.GetUserNameByIdRequest{
		UserId: userID.String(),
	})
	if err != nil {
		s.Log.Error("error with getting user name: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	topicName, err := s.Repository.SelectTopicName(ctx.Request.Context(), topicID, body.ChannelId)
	if err != nil {
		s.Log.Error("error with getting ")
	}

	s.Hub.onSendMessage(dto.AddMessage{
		ChatID:      body.ChannelId,
		SenderID:    uuid.Nil,
		Message:     fmt.Sprintf("User %s deleted the topic name %s", response.GetName(), topicName),
		MessageType: "system",
	})

	ctx.JSON(http.StatusOK, topicID)
}
