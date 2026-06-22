package servicehttp

import (
	"errors"
	"net/http"
	"time"

	"github.com/Flikest/PingVi_backend/internal/delivery/dto"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (s *ServiceMessenger) GetPermissions(ctx *gin.Context) {
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

	permission, err := s.Repository.SelectPermissions(ctx.Request.Context(), userID, channelID)
	if err != nil {
		s.Log.Error("error with getting user permission from channel: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, permission)
}

func (s *ServiceMessenger) CreateRole(ctx *gin.Context) {
	var body dto.CreateRole
	if err := ctx.BindJSON(&body); err != nil {
		s.Log.Error("invalid body: ", "error", err)
		ctx.JSON(http.StatusBadRequest, err)
		return
	}

	// Бизнес-логика: проверка прав
	permissions, err := s.Repository.SelectPermissions(ctx.Request.Context(), body.UserID, body.ChannelID)
	if err != nil {
		s.Log.Error("error with selecting permissions: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}

	if len(permissions) < 5 || permissions[4] != '1' {
		s.Log.Error("the user does not have sufficient rights")
		ctx.JSON(http.StatusForbidden, errors.New("the user does not have sufficient rights"))
		return
	}

	// Бизнес-логика: генерация ID и времени
	roleID, err := uuid.NewV7()
	if err != nil {
		s.Log.Error("error with generating role id: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}

	now := time.Now()

	// Создание роли
	role := dto.Role{
		ID:            roleID,
		ChannelID:     body.ChannelID,
		Name:          body.Name,
		Color:         body.Color,
		Permissions:   body.Permissions,
		IsMentionable: body.IsMentionable,
		IsDefault:     body.IsDefault,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if err := s.Repository.InsertChannelRole(ctx.Request.Context(), role); err != nil {
		s.Log.Error("error with inserting role", "error", err)
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}

	// TODO добавить рассылку о том что пользователь создал роль

	ctx.JSON(http.StatusCreated, role)
}

func (s *ServiceMessenger) GetAllRoles(ctx *gin.Context) {
	channelID, err := uuid.Parse(ctx.Param("channel_id"))
	if err != nil {
		s.Log.Error("error with parsing user id uuid: ", "error", err)
		ctx.JSON(http.StatusBadRequest, err)
		return
	}

	roles, err := s.Repository.SelectRoles(ctx.Request.Context(), channelID)
	if err != nil {
		s.Log.Error("error with getting roles from channel: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, roles)
}

func (s *ServiceMessenger) UpdateRole(ctx *gin.Context) {
	var body dto.UpdateRole
	if err := ctx.BindJSON(&body); err != nil {
		s.Log.Error("invalid body: ", "error", err)
		ctx.JSON(http.StatusBadRequest, err)
		return
	}

	// Бизнес-логика: проверка прав
	permissions, err := s.Repository.SelectPermissions(ctx.Request.Context(), body.UserID, body.ChannelID)
	if err != nil {
		s.Log.Error("error with selecting permissions: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}

	if permissions[4] != '1' {
		s.Log.Error("the user does not have sufficient rights")
		ctx.JSON(http.StatusForbidden, errors.New("the user does not have sufficient rights"))
		return
	}

	role, err := s.Repository.UpdateRole(ctx.Request.Context(), body)
	if err != nil {
		s.Log.Error("error with updating role: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, err.Error())
		return

	}

	ctx.JSON(http.StatusOK, role)
}

func (s *ServiceMessenger) DeleteRole(ctx *gin.Context) {
	var body dto.DeleteRole
	if err := ctx.BindJSON(&body); err != nil {
		s.Log.Error("invalid body: ", "error", err)
		ctx.JSON(http.StatusBadRequest, err)
		return
	}

	// Бизнес-логика: проверка прав
	permissions, err := s.Repository.SelectPermissions(ctx.Request.Context(), body.UserID, body.ChannelID)
	if err != nil {
		s.Log.Error("error with selecting permissions: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}

	if permissions[4] != '1' {
		s.Log.Error("the user does not have sufficient rights")
		ctx.JSON(http.StatusForbidden, errors.New("the user does not have sufficient rights"))
		return
	}

	// Получение роли @everyone
	everyoneRoleID, err := s.Repository.SelectEveryoneRoleID(ctx.Request.Context(), body.ChannelID)
	if err != nil {
		s.Log.Error("error with selecting role id on everyone name: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}

	userIDs, err := s.Repository.SelectMemberIDsByRoleID(ctx.Request.Context(), body.ChannelID, body.RoleID)
	if err != nil {
		s.Log.Error("error with selecting member ids by role id: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}

	if err := s.Repository.UpdateChannelMembersRole(ctx.Request.Context(), body.ChannelID, everyoneRoleID, userIDs); err != nil {
		s.Log.Error("error with updating members role: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}

	if err := s.Repository.DeleteChannelRoleByID(ctx.Request.Context(), body.RoleID, body.ChannelID); err != nil {
		s.Log.Error("error with deleting role: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}

	// TODO добавить рассылку о том что пользователь удалил роль канала

	ctx.JSON(http.StatusOK, body.RoleID)
}
