package servicehttp

import (
	"errors"
	"net/http"
	"time"

	pb "github.com/Flikest/PingVi_backend/gen/go/user_info"
	"github.com/Flikest/PingVi_backend/internal/delivery/dto"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (s *ServiceMessenger) CreateChannel(ctx *gin.Context) {
	var body dto.CreateChannel
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
		OwnerID:     body.OwnerID,
		IconURL:     body.IconURL,
		IsPublic:    body.IsPublic,
		CreatedAt:   now,
		UpdatedAt:   now,
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
		UserID:    body.OwnerID,
		RoleID:    ownerRoleID,
		JoinedAt:  now,
	}

	if err := s.Repository.InsertChannelMember(ctx.Request.Context(), ownerMember); err != nil {
		s.Log.Error("error with inserting owner member: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}

	_, err = s.Client.GetUserNameByID(ctx.Request.Context(), &pb.GetUserNameByIdRequest{UserId: body.OwnerID.String()})
	if err != nil {
		s.Log.Error("error getting user name: ", "error", err)
	}

	// TODO: создать топик и директорию по general
	// TODO добавить рассылку о том что канал, топик и вкладка был создан

	ctx.JSON(http.StatusCreated, channel)
}

func (s *ServiceMessenger) JoinChannel(ctx *gin.Context) {
	userID, err := uuid.Parse(ctx.Param("user_id"))
	if err != nil {
		s.Log.Error("error with parsing user id uuid: ", "error", err)
		ctx.JSON(http.StatusBadRequest, err)
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

	// TODO добавить рассылку о том что пользователь зашел в канал

	ctx.JSON(http.StatusOK, channelID)
}
func (s *ServiceMessenger) CickChannelMember(ctx *gin.Context) {
	var body dto.CickChannelMember
	if err := ctx.BindJSON(&body); err != nil {
		s.Log.Error("invalid body: ", "error", err)
		ctx.JSON(http.StatusBadRequest, err)
		return
	}

	permission, err := s.Repository.SelectPermissions(ctx.Request.Context(), body.MemberID, body.ChannelID)
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

	// Удаление участника
	if err := s.Repository.DeleteChannelMemberByUserID(ctx.Request.Context(), body.ChannelID, body.KickedMemberID); err != nil {
		s.Log.Error("error with deleting member from channel: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}

	// TODO добавить рассылку о том что пользователь был кикнут

	ctx.JSON(http.StatusOK, body.KickedMemberID)
}

func (s *ServiceMessenger) LeaveFromChannel(ctx *gin.Context) {
	channelID, err := uuid.Parse(ctx.Param("channel_id"))
	if err != nil {
		s.Log.Error("error with parsing user id uuid: ", "error", err)
		ctx.JSON(http.StatusBadRequest, err)
		return
	}

	userID, err := uuid.Parse(ctx.Param("user_id"))
	if err != nil {
		s.Log.Error("error with parsing user id uuid: ", "error", err)
		ctx.JSON(http.StatusBadRequest, err)
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

		// TODO добавить рассылку о том что канал был удален

		ctx.JSON(http.StatusOK, channelID)
		return
	}

	if err := s.Repository.DeleteChannelMemberByUserID(ctx.Request.Context(), channelID, userID); err != nil {
		s.Log.Error("error with delete channel member: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, err.Error())
		return
	}

	// TODO добавить рассылку о том что пользователь вышел из чата

	ctx.JSON(http.StatusOK, channelID)
}

func (s *ServiceMessenger) GetChannels(ctx *gin.Context) {
	userID, err := uuid.Parse(ctx.Param("user_id"))
	if err != nil {
		s.Log.Error("error with parsing user id uuid: ", "error", err)
		ctx.JSON(http.StatusBadRequest, err)
		return
	}

	// Получение каналов пользователя
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
	var body dto.UpdateChannel
	if err := ctx.BindJSON(&body); err != nil {
		s.Log.Error("invalid body: ", "error", err)
		ctx.JSON(http.StatusBadRequest, err)
		return
	}

	permissions, err := s.Repository.SelectPermissions(ctx.Request.Context(), body.UserID, body.ID)
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

	// TODO добавить рассылку в канал о том что пользователь изменил канал

	ctx.JSON(http.StatusOK, channel)
}

func (s *ServiceMessenger) Deletechannel(ctx *gin.Context) {
	var body dto.DeleteChannel
	if err := ctx.BindJSON(&body); err != nil {
		s.Log.Error("invalid body: ", "error", err)
		ctx.JSON(http.StatusBadRequest, err)
		return
	}

	permissions, err := s.Repository.SelectPermissions(ctx.Request.Context(), body.UserID, body.ID)
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

	// TODO добавить рассылку о том что канал был удален

	ctx.JSON(http.StatusOK, body.ID)
}
