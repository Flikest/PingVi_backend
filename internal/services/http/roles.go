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

	var body dto.CreateRole
	if err := ctx.BindJSON(&body); err != nil {
		s.Log.Error("invalid body: ", "error", err)
		ctx.JSON(http.StatusBadRequest, err)
		return
	}

	body.UserID = userID

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

	roleID, err := uuid.NewV7()
	if err != nil {
		s.Log.Error("error with generating role id: ", "error", err)
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

	now := time.Now()

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

	s.Hub.onSendMessage(dto.AddMessage{
		ChatID:      body.ChannelID,
		SenderID:    uuid.Nil,
		Message:     fmt.Sprintf("User %s created the role %s", response.GetName(), role.Name),
		MessageType: "system",
	})

	ctx.JSON(http.StatusCreated, role)
}

func (s *ServiceMessenger) GetAllRoles(ctx *gin.Context) {
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

	isBelong, err := s.Repository.IsBelongGroup(ctx.Request.Context(), channelID, userID)

	if !isBelong {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "the user does not belong to this group"})
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

	var body dto.UpdateRole
	if err := ctx.BindJSON(&body); err != nil {
		s.Log.Error("invalid body: ", "error", err)
		ctx.JSON(http.StatusBadRequest, err)
		return
	}

	body.UserID = userID

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

	var body dto.DeleteRole
	if err := ctx.BindJSON(&body); err != nil {
		s.Log.Error("invalid body: ", "error", err)
		ctx.JSON(http.StatusBadRequest, err)
		return
	}

	body.UserID = userID

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

	roleName, err := s.Repository.SelectRoleName(ctx.Request.Context(), body.RoleID, body.ChannelID)
	if err != nil {
		s.Log.Error("error with getting role name: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

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

	response, err := s.Client.GetUserNameByID(ctx.Request.Context(), &pb.GetUserNameByIdRequest{
		UserId: userID.String(),
	})
	if err != nil {
		s.Log.Error("error with getting user name: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	s.Hub.onSendMessage(dto.AddMessage{
		ChatID:      body.ChannelID,
		SenderID:    uuid.Nil,
		Message:     fmt.Sprintf("User %s deleted the role %s", response.GetName(), roleName),
		MessageType: "system",
	})

	ctx.JSON(http.StatusOK, body.RoleID)
}
