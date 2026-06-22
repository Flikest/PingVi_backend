package servicehttp

import (
	"errors"
	"net/http"
	"time"

	"github.com/Flikest/PingVi_backend/internal/delivery/dto"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (s *ServiceMessenger) CreateGroup(ctx *gin.Context) {
	var body dto.CreateGroup
	if err := ctx.BindJSON(&body); err != nil {
		s.Log.Error("invalid body: ", "error", err)
		ctx.JSON(http.StatusBadRequest, err)
		return
	}

	// Бизнес-логика: генерация ID и времени
	groupID, err := uuid.NewV7()
	if err != nil {
		s.Log.Error("error with generating group id: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}

	now := time.Now()

	group := dto.Group{
		ID:          groupID,
		Name:        body.Name,
		AvatarLinks: body.AvatarLinks,
		OwnerID:     body.OwnerID,
		IsPublic:    body.IsPublic,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.Repository.InsertGroup(ctx.Request.Context(), group); err != nil {
		s.Log.Error("error with creating group: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}

	// TODO добавить рассылку сообщения о том что группа была создана

	ctx.JSON(http.StatusCreated, group)
}

func (s *ServiceMessenger) JoinGroup(ctx *gin.Context) {
	groupID, err := uuid.Parse(ctx.Param("group_id"))
	if err != nil {
		s.Log.Error("error with parsing group id uuid", "error", err)
		ctx.JSON(http.StatusBadRequest, err)
		return
	}

	userID, err := uuid.Parse(ctx.Param("user_id"))
	if err != nil {
		s.Log.Error("error with parse user id uuid: ", "error", err)
		ctx.JSON(http.StatusBadRequest, err)
		return
	}

	// Добавление участника в группу
	member := dto.GroupMember{
		GroupID:  groupID,
		UserID:   userID,
		IsAdmin:  false,
		JoinedAt: time.Now(),
	}

	if err := s.Repository.InsertGroupMember(ctx.Request.Context(), member); err != nil {
		s.Log.Error("error when user logs into group: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}

	// TODO добавить отправку сообщения в группу о входе пользователя

	ctx.JSON(http.StatusOK, groupID)
}

func (s *ServiceMessenger) LeaveGroup(ctx *gin.Context) {
	groupID, err := uuid.Parse(ctx.Param("group_id"))
	if err != nil {
		s.Log.Error("error with parsing group id uuid", "error", err)
		ctx.JSON(http.StatusBadRequest, err)
		return
	}

	userID, err := uuid.Parse(ctx.Param("user_id"))
	if err != nil {
		s.Log.Error("error with parse user id uuid: ", "error", err)
		ctx.JSON(http.StatusBadRequest, err)
		return
	}

	// Бизнес-логика: проверка прав
	ownerID, err := s.Repository.SelectGroupOwnerID(ctx.Request.Context(), groupID)
	if err != nil {
		s.Log.Error("error with checking owner: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}

	if ownerID == userID {
		// Владелец удаляет группу полностью
		// Удаление всех участников
		if err := s.Repository.DeleteGroupMembersByGroupID(ctx.Request.Context(), groupID); err != nil {
			s.Log.Error("error with deleting members", "error", err)
			ctx.JSON(http.StatusInternalServerError, err)
			return
		}

		// Удаление группы
		if err := s.Repository.DeleteGroupByID(ctx.Request.Context(), groupID); err != nil {
			s.Log.Error("error with deleting group", "error", err)
			ctx.JSON(http.StatusInternalServerError, err)
			return
		}

		// TODO добавить рассылку о том что группа была удалена
		ctx.JSON(http.StatusOK, groupID)
		return
	}

	// Обычный участник просто выходит
	if err := s.Repository.DeleteGroupMember(ctx.Request.Context(), groupID, userID); err != nil {
		s.Log.Error("error when user leaves group: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}

	// TODO добавить рассылку сообщен��й в группу о выходе пользователя из группы
	ctx.JSON(http.StatusOK, userID)
}

func (s *ServiceMessenger) SelectAllMembersGroup(ctx *gin.Context) {
	groupID, err := uuid.Parse(ctx.Param("group_id"))
	if err != nil {
		s.Log.Error("error with parsing group id uuid", "error", err)
		ctx.JSON(http.StatusBadRequest, err)
		return
	}

	groupMembers, err := s.Repository.SelectAllMembersGroup(ctx.Request.Context(), groupID)
	if err != nil {
		s.Log.Error("error with getting all members from group")
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, groupMembers)
}

func (s *ServiceMessenger) SelectMemberGroup(ctx *gin.Context) {
	groupID, err := uuid.Parse(ctx.Param("group_id"))
	if err != nil {
		s.Log.Error("error with parsing group id uuid", "error", err)
		ctx.JSON(http.StatusBadRequest, err)
		return
	}

	userID, err := uuid.Parse(ctx.Param("user_id"))
	if err != nil {
		s.Log.Error("error with parse user id uuid: ", "error", err)
		ctx.JSON(http.StatusBadRequest, err)
		return
	}

	member, err := s.Repository.SelectMemberGroup(ctx.Request.Context(), groupID, userID)
	if err != nil {
		s.Log.Error("error with getting member from group: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, member)
}

func (s *ServiceMessenger) KickMemberFromGroup(ctx *gin.Context) {
	var body dto.CickGroupMember
	if err := ctx.BindJSON(&body); err != nil {
		s.Log.Error("invalid body: ", "error", err)
		ctx.JSON(http.StatusBadRequest, err)
		return
	}

	// Бизнес-логика: проверка прав
	kickerIsAdmin, err := s.Repository.SelectGroupMemberIsAdmin(ctx.Request.Context(), body.GroupID, body.MemberID)
	if err != nil {
		s.Log.Error("error checking permissions: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}

	ownerID, err := s.Repository.SelectGroupOwnerID(ctx.Request.Context(), body.GroupID)
	if err != nil {
		s.Log.Error("error checking permissions: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}

	if ownerID == body.MemberID {
		s.Log.Error("You can't kick the owner")
		ctx.JSON(http.StatusForbidden, errors.New("You can't kick the owner"))
		return
	}

	if !kickerIsAdmin && ownerID != body.MemberID {
		s.Log.Warn("not enough rights")
		ctx.JSON(http.StatusForbidden, errors.New("not enough rights"))
		return
	}

	// Удаление участника
	if err := s.Repository.DeleteGroupMember(ctx.Request.Context(), body.GroupID, body.MemberID); err != nil {
		s.Log.Error("error with deleting member from group: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, body.GroupID)
}

func (s *ServiceMessenger) SelectGroup(ctx *gin.Context) {
	userID, err := uuid.Parse(ctx.Param("user_id"))
	if err != nil {
		s.Log.Error("error with parse user id uuid: ", "error", err)
		ctx.JSON(http.StatusBadRequest, err)
		return
	}

	groups, err := s.Repository.SelectGroup(ctx.Request.Context(), userID)
	if err != nil {
		s.Log.Error("Error searching for all user groups", "error", err)
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, groups)
}

func (s *ServiceMessenger) UpdateGroup(ctx *gin.Context) {
	var body dto.UpdateGroup
	if err := ctx.BindJSON(&body); err != nil {
		s.Log.Error("invalid body: ", "error", err)
		ctx.JSON(http.StatusBadRequest, err)
		return
	}

	// Бизнес-логика: проверка прав
	isAdmin, err := s.Repository.SelectGroupMemberIsAdmin(ctx.Request.Context(), body.ID, body.UserID)
	if err != nil {
		s.Log.Error("error with checking permissions: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}

	ownerID, err := s.Repository.SelectGroupOwnerID(ctx.Request.Context(), body.ID)
	if err != nil {
		s.Log.Error("error with checking permissions: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}

	if !isAdmin && ownerID != body.UserID {
		s.Log.Warn("not enough rights")
		ctx.JSON(http.StatusForbidden, errors.New("not enough rights"))
		return
	}

	// Проверка существования группы
	exists, err := s.Repository.SelectGroupExists(ctx.Request.Context(), body.ID)
	if err != nil {
		s.Log.Error("error with selecting group: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}

	if !exists {
		ctx.JSON(http.StatusNotFound, errors.New("such a group does not exist"))
		return
	}

	now := time.Now()

	// Обновление группы
	group := dto.Group{
		ID:          body.ID,
		Name:        body.Name,
		AvatarLinks: body.AvatarLinks,
		OwnerID:     body.OwnerID,
		IsPublic:    body.IsPublic,
		CreatedAt:   body.CreatedAt,
		UpdatedAt:   now,
	}

	if err := s.Repository.UpdateGroup(ctx.Request.Context(), group); err != nil {
		s.Log.Error("error with updating group: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}

	// TODO добавить рассылку в группу о том что пользователь изменил группу

	ctx.JSON(http.StatusOK, group)
}

func (s *ServiceMessenger) DeleteGroup(ctx *gin.Context) {
	var body dto.DeleteGroup
	if err := ctx.BindJSON(&body); err != nil {
		s.Log.Error("invalid body: ", "error", err)
		ctx.JSON(http.StatusBadRequest, err)
		return
	}

	// Бизнес-логика: проверка прав
	isAdmin, err := s.Repository.SelectGroupMemberIsAdmin(ctx.Request.Context(), body.ID, body.UserID)
	if err != nil {
		s.Log.Error("error with checking permissions: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}

	ownerID, err := s.Repository.SelectGroupOwnerID(ctx.Request.Context(), body.ID)
	if err != nil {
		s.Log.Error("error with checking permissions: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}

	if !isAdmin && ownerID != body.UserID {
		s.Log.Warn("not enough rights")
		ctx.JSON(http.StatusForbidden, errors.New("not enough rights"))
		return
	}

	// Удаление всех участников
	if err := s.Repository.DeleteGroupMembersByGroupID(ctx.Request.Context(), body.ID); err != nil {
		s.Log.Error("error with deleting members: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}

	// Удаление группы
	if err := s.Repository.DeleteGroupByID(ctx.Request.Context(), body.ID); err != nil {
		s.Log.Error("error with deleting group", "error", err)
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}

	// TODO добавить рассылку о том что группа была удалена

	ctx.JSON(http.StatusOK, body.ID)
}
