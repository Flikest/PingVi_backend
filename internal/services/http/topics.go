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
