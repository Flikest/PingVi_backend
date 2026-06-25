package servicehttp

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	pb "github.com/Flikest/PingVi_backend/gen/go/user_info"
	"github.com/Flikest/PingVi_backend/internal/delivery/dto"
	"github.com/Flikest/PingVi_backend/pkg/tokens"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (s *ServiceMessenger) CreateChannel(ctx *gin.Context) {
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

	var body dto.CreateChannelRequest
	if err := ctx.BindJSON(&body); err != nil {
		s.Log.Error("invalid body: ", "error", err)
		ctx.JSON(http.StatusBadRequest, err)
		return
	}

	channelID, err := uuid.NewV7()
	if err != nil {
		s.Log.Error("error with generating channel id: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}

	ownerRoleID, err := uuid.NewV7()
	if err != nil {
		s.Log.Error("error with generating role id: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}

	now := time.Now()

	channel := dto.Channel{
		ID:          channelID,
		Name:        body.Name,
		Description: body.Description,
		OwnerID:     userID,
		IconURL:     body.IconURL,
		IsPublic:    body.IsPublic,
		CreatedAt:   now,
	}

	if err := s.Repository.InsertChannel(ctx.Request.Context(), channel); err != nil {
		s.Log.Error("error with inserting channel: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}

	ownerRole := dto.Role{
		ID:            ownerRoleID,
		ChannelID:     channelID,
		Name:          "@owner",
		Color:         "#FF0000",
		Permissions:   "111111",
		IsMentionable: false,
		IsDefault:     false,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := s.Repository.InsertChannelRole(ctx.Request.Context(), ownerRole); err != nil {
		s.Log.Error("error with inserting owner role in chat: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}

	ownerMember := dto.ChannelMember{
		ChannelID: channelID,
		UserID:    userID,
		RoleID:    ownerRoleID,
		JoinedAt:  now,
	}

	if err := s.Repository.InsertChannelMember(ctx.Request.Context(), ownerMember); err != nil {
		s.Log.Error("error with inserting owner member: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}

	_, err = s.Client.GetUserNameByID(ctx.Request.Context(), &pb.GetUserNameByIdRequest{UserId: userID.String()})
	if err != nil {
		s.Log.Error("error getting user name: ", "error", err)
	}

	// TODO: создать топик и директорию по general

	response, err := s.Client.GetUserNameByID(ctx.Request.Context(), &pb.GetUserNameByIdRequest{
		UserId: userID.String(),
	})
	if err != nil {
		s.Log.Error("error with getting user name: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	s.Hub.onSendMessage(dto.AddMessage{
		ChatID:      channelID,
		SenderID:    uuid.Nil,
		Message:     fmt.Sprintf("User %s create this channel", response.GetName()),
		MessageType: "system",
	})

	s.Hub.onSendMessage(dto.AddMessage{
		ChatID:      channelID,
		SenderID:    uuid.Nil,
		Message:     "General topic was created",
		MessageType: "system",
	})

	s.Hub.onSendMessage(dto.AddMessage{
		ChatID:      channelID,
		SenderID:    uuid.Nil,
		Message:     "General directory was created",
		MessageType: "system",
	})

	ctx.JSON(http.StatusCreated, channel)
}

func (s *ServiceMessenger) JoinChannel(ctx *gin.Context) {
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

	everyoneRoleID, err := s.Repository.SelectEveryoneRoleID(ctx.Request.Context(), channelID)
	if err != nil {
		s.Log.Error("error with selecting everyone role id: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}

	member := dto.ChannelMember{
		ChannelID: channelID,
		UserID:    userID,
		RoleID:    everyoneRoleID,
		JoinedAt:  time.Now(),
	}

	if err := s.Repository.InsertChannelMember(ctx.Request.Context(), member); err != nil {
		s.Log.Error("error with inserting member: ", "error", err)
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

	s.Hub.onSendMessage(dto.AddMessage{
		ChatID:      channelID,
		SenderID:    uuid.Nil,
		Message:     fmt.Sprintf("User %s joined the channel", response.GetName()),
		MessageType: "system",
	})

	ctx.JSON(http.StatusOK, channelID)
}
func (s *ServiceMessenger) CickChannelMember(ctx *gin.Context) {
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

	var bodyRequest dto.CickChannelMemberRequest
	if err := ctx.BindJSON(&bodyRequest); err != nil {
		s.Log.Error("invalid body: ", "error", err)
		ctx.JSON(http.StatusBadRequest, err)
		return
	}

	permission, err := s.Repository.SelectPermissions(ctx.Request.Context(), userID, bodyRequest.ChannelID)
	if err != nil {
		s.Log.Error("error with selecting user permission from channel: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}

	if permission[0] != '1' {
		s.Log.Error("the user does not have enough rights")
		ctx.JSON(http.StatusForbidden, errors.New("the user does not have enough rights"))
		return
	}

	if err := s.Repository.DeleteChannelMemberByUserID(ctx.Request.Context(), bodyRequest.ChannelID, bodyRequest.KickedMemberID); err != nil {
		s.Log.Error("error with deleting member from channel: ", "error", err)
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

	s.Hub.onSendMessage(dto.AddMessage{
		ChatID:      bodyRequest.ChannelID,
		SenderID:    uuid.Nil,
		Message:     fmt.Sprintf("User %s was kicked from the channel", response.GetName()),
		MessageType: "system",
	})

	ctx.JSON(http.StatusOK, bodyRequest.KickedMemberID)
}

func (s *ServiceMessenger) LeaveFromChannel(ctx *gin.Context) {
	channelID, err := uuid.Parse(ctx.Param("channel_id"))
	if err != nil {
		s.Log.Error("error with parsing user id uuid: ", "error", err)
		ctx.JSON(http.StatusBadRequest, err)
		return
	}

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

	ownerID, err := s.Repository.SelectChannelOwnerID(ctx.Request.Context(), channelID)
	if err != nil {
		s.Log.Error("error with selecting owner id: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	if ownerID == userID {
		permissions, err := s.Repository.SelectPermissions(ctx.Request.Context(), userID, channelID)
		if err != nil {
			s.Log.Error("error with selecting user permissions from channel: ", "error", err)
			ctx.JSON(http.StatusInternalServerError, err)
			return
		}

		if permissions[2] != '1' {
			s.Log.Error("the user does not have enough rights")
			ctx.JSON(http.StatusForbidden, errors.New("the user does not have enough rights"))
			return
		}

		if err := s.Repository.DeleteChannelRolesByChannelID(ctx.Request.Context(), channelID); err != nil {
			s.Log.Error("error with deleting roles from chat: ", "error", err)
			ctx.JSON(http.StatusInternalServerError, err)
			return
		}

		if err := s.Repository.DeleteChannelTopicsByChannelID(ctx.Request.Context(), channelID); err != nil {
			s.Log.Error("error with deleting topics from channel: ", "error", err)
			ctx.JSON(http.StatusInternalServerError, err)
			return
		}

		if err := s.Repository.DeleteChannelMembersByChannelID(ctx.Request.Context(), channelID); err != nil {
			s.Log.Error("error with deleting members from chat: ", "error", err)
			ctx.JSON(http.StatusInternalServerError, err)
			return
		}

		if err := s.Repository.DeleteChannelByID(ctx.Request.Context(), channelID); err != nil {
			s.Log.Error("error with deleting channel: ", "error", err)
			ctx.JSON(http.StatusInternalServerError, err)
			return
		}

		channelName, err := s.Repository.SelectChannelNameByID(ctx.Request.Context(), channelID)
		if err != nil {
			s.Log.Error("error with selecting channel name: ", "error", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		s.Hub.onSendMessage(dto.AddMessage{
			ChatID:      channelID,
			SenderID:    uuid.Nil,
			Message:     fmt.Sprintf("channel %s was deleted", channelName),
			MessageType: "system",
		})

		ctx.JSON(http.StatusOK, channelID)
		return
	}

	if err := s.Repository.DeleteChannelMemberByUserID(ctx.Request.Context(), channelID, userID); err != nil {
		s.Log.Error("error with delete channel member: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, err.Error())
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

	s.Hub.onSendMessage(dto.AddMessage{
		ChatID:      channelID,
		SenderID:    uuid.Nil,
		Message:     fmt.Sprintf("User %s has left the channel", response.GetName()),
		MessageType: "system",
	})

	ctx.JSON(http.StatusOK, channelID)
}

func (s *ServiceMessenger) GetChannels(ctx *gin.Context) {
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

	selectMembers := s.Repository.Session.ContextQuery(
		ctx.Request.Context(),
		`SELECT channel_id FROM messenger_keyspace.channel_members WHERE user_id IN ?`,
		[]string{":user_id"}).
		BindMap(map[string]interface{}{
			":user_id": userID,
		})

	var userIDs []string
	if err := selectMembers.SelectRelease(&userIDs); err != nil {
		s.Log.Error("error with selecting members by id: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}

	query := s.Repository.Session.Query(
		`SELECT * FROM messenger_keyspace.channels WHERE id=?`,
		[]string{":id"}).
		BindMap(map[string]interface{}{
			":id": userIDs,
		})

	var channels []dto.Channel
	if err := query.SelectRelease(&channels); err != nil {
		s.Log.Error("error with selecting channels: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, channels)
}

func (s *ServiceMessenger) UpdateChannel(ctx *gin.Context) {
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

	var body dto.UpdateChannelRequest
	if err := ctx.BindJSON(&body); err != nil {
		s.Log.Error("invalid body: ", "error", err)
		ctx.JSON(http.StatusBadRequest, err)
		return
	}

	permissions, err := s.Repository.SelectPermissions(ctx.Request.Context(), userID, body.ID)
	if err != nil {
		s.Log.Error("error with selecting user permissions from channel: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}

	if permissions[2] != '1' {
		s.Log.Error("the user does not have enough rights")
		ctx.JSON(http.StatusForbidden, errors.New("the user does not have enough rights"))
		return
	}

	now := time.Now()

	channel := dto.Channel{
		ID:          body.ID,
		Name:        body.Name,
		Description: body.Description,
		OwnerID:     body.OwnerID,
		IconURL:     body.IconURL,
		IsPublic:    body.IsPublic,
		CreatedAt:   body.CreatedAt,
		UpdatedAt:   now,
	}

	if err := s.Repository.UpdateChannel(ctx.Request.Context(), channel); err != nil {
		s.Log.Error("error with updating channel: ", "error", err)
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

	s.Hub.onSendMessage(dto.AddMessage{
		ChatID:      channel.ID,
		SenderID:    uuid.Nil,
		Message:     fmt.Sprintf("User %s changed the group", response.GetName()),
		MessageType: "system",
	})

	ctx.JSON(http.StatusOK, channel)
}

func (s *ServiceMessenger) Deletechannel(ctx *gin.Context) {
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

	var body dto.DeleteChannelRequest
	if err := ctx.BindJSON(&body); err != nil {
		s.Log.Error("invalid body: ", "error", err)
		ctx.JSON(http.StatusBadRequest, err)
		return
	}

	permissions, err := s.Repository.SelectPermissions(ctx.Request.Context(), userID, body.ID)
	if err != nil {
		s.Log.Error("error with selecting user permissions from channel: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}

	if permissions[2] != '1' {
		s.Log.Error("the user does not have enough rights")
		ctx.JSON(http.StatusForbidden, errors.New("the user does not have enough rights"))
		return
	}

	if err := s.Repository.DeleteChannelRolesByChannelID(ctx.Request.Context(), body.ID); err != nil {
		s.Log.Error("error with deleting roles from chat: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}

	if err := s.Repository.DeleteChannelTopicsByChannelID(ctx.Request.Context(), body.ID); err != nil {
		s.Log.Error("error with deleting topics from channel: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}

	if err := s.Repository.DeleteChannelMembersByChannelID(ctx.Request.Context(), body.ID); err != nil {
		s.Log.Error("error with deleting members from chat: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}

	if err := s.Repository.DeleteChannelByID(ctx.Request.Context(), body.ID); err != nil {
		s.Log.Error("error with deleting channel: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}

	channelName, err := s.Repository.SelectChannelNameByID(ctx.Request.Context(), body.ID)
	if err != nil {
		s.Log.Error("error with selecting channel name: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	s.Hub.onSendMessage(dto.AddMessage{
		ChatID:      body.ID,
		SenderID:    uuid.Nil,
		Message:     fmt.Sprintf("channel %s was deleted", channelName),
		MessageType: "system",
	})

	ctx.JSON(http.StatusOK, body.ID)
}
