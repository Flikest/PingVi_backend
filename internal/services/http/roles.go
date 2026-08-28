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

// GetPermissions godoc
//
//	@Summary		Get user permissions in a channel
//	@Description	Retrieves the permission string for a user in a specific channel.
//	@Description	- Permission string format: 6 characters (e.g., "111111")
//	@Description	- Each position represents a different permission
//	@Tags			messenger-roles
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			Authorization	header		string					true					"Bearer JWT token"	example("Bearer eyJhbGciOiJIUzI1NiIs...")
//	@Param			channel_id		path		string					true					"Channel ID"		example("123e4567-e89b-12d3-a456-426614174000")
//	@Param			user_id			path		string					true					"User ID"			example("987fcdeb-51d2-12d3-a456-426614174000")
//	@Success		200				{string}	string					"Permission string"		example("111111")
//	@Failure		400				{object}	map[string]interface{}	"Invalid IDs"			example({"error":"invalid user id"})
//	@Failure		500				{object}	map[string]interface{}	"Internal server error"	example({"error":"failed to get permissions"})
//	@Router			/roles/permissions/{channel_id} [get]
func (s *ServiceMessenger) GetPermissions(ctx *gin.Context) {
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

	permission, err := s.Repository.SelectPermissions(ctx.Request.Context(), userID, channelID)
	if err != nil {
		s.Log.Error("error with getting user permission from channel: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, permission)
}

// CreateRole godoc
//
//	@Summary		Create a new role in a channel
//	@Description	Creates a new role in a channel. Requires 'manage_roles' permission (permission[4] == '1').
//	@Description	- Can set: name, color, permissions, mentionable flag
//	@Description	- System message will be sent to the channel
//	@Tags			messenger-roles
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			Authorization	header		string					true	"Bearer JWT token"		example("Bearer eyJhbGciOiJIUzI1NiIs...")
//	@Param			request			body		dto.CreateRole			true	"Role creation data"	example({"channel_id":"123e4567-e89b-12d3-a456-426614174000","name":"Moderator","color":"#00FF00","permissions":"101000","is_mentionable":true,"is_default":false})
//	@Success		201				{object}	dto.Role				"Created role"
//	@Failure		400				{object}	map[string]interface{}	"Invalid request"			example({"error":"invalid body"})
//	@Failure		403				{object}	map[string]interface{}	"Insufficient permissions"	example({"error":"the user does not have sufficient rights"})
//	@Failure		500				{object}	map[string]interface{}	"Internal server error"		example({"error":"failed to create role"})
//	@Router			/roles [post]
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

	s.Hub.onSendMessage(ctx.Request.Context(), dto.AddMessage{
		ChatID:   body.ChannelID,
		SenderID: uuid.Nil,
		Message:  fmt.Sprintf("User %s created the role %s", response.GetName(), role.Name),
		IsSystem: true,
	})

	ctx.JSON(http.StatusCreated, role)
}

// GetAllRoles godoc
//
//	@Summary		Get all roles in a channel
//	@Description	Retrieves all roles from a specific channel. User must be a member of the channel.
//	@Tags			messenger-roles
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			Authorization	header		string					true	"Bearer JWT token"	example("Bearer eyJhbGciOiJIUzI1NiIs...")
//	@Param			channel_id		path		string					true	"Channel ID"		example("123e4567-e89b-12d3-a456-426614174000")
//	@Success		200				{array}		dto.Role				"List of roles"
//	@Failure		400				{object}	map[string]interface{}	"Invalid channel ID"	example({"error":"invalid channel id"})
//	@Failure		403				{object}	map[string]interface{}	"Not a member"			example({"error":"the user does not belong to this group"})
//	@Failure		500				{object}	map[string]interface{}	"Internal server error"	example({"error":"failed to get roles"})
//	@Router			/roles/{channel_id} [get]
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

// UpdateRole godoc
//
//	@Summary		Update a role
//	@Description	Updates an existing role in a channel. Requires 'manage_roles' permission (permission[4] == '1').
//	@Description	- Can update: name, color, permissions, mentionable flag
//	@Tags			messenger-roles
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			Authorization	header		string					true	"Bearer JWT token"	example("Bearer eyJhbGciOiJIUzI1NiIs...")
//	@Param			request			body		dto.UpdateRole			true	"Role update data"	example({"id":"123e4567-e89b-12d3-a456-426614174000","channel_id":"123e4567-e89b-12d3-a456-426614174000","name":"New Moderator","color":"#FF0000","permissions":"111000","is_mentionable":true})
//	@Success		200				{object}	dto.Role				"Updated role"
//	@Failure		400				{object}	map[string]interface{}	"Invalid request"			example({"error":"invalid body"})
//	@Failure		403				{object}	map[string]interface{}	"Insufficient permissions"	example({"error":"the user does not have sufficient rights"})
//	@Failure		500				{object}	map[string]interface{}	"Internal server error"		example({"error":"failed to update role"})
//	@Router			/roles [put]
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

// DeleteRole godoc
//
//	@Summary		Delete a role
//	@Description	Permanently deletes a role from a channel. Members with this role will be reassigned to @everyone role.
//	@Description	- Requires 'manage_roles' permission (permission[4] == '1')
//	@Description	- System message will be sent to the channel
//	@Tags			messenger-roles
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			Authorization	header		string					true						"Bearer JWT token"		example("Bearer eyJhbGciOiJIUzI1NiIs...")
//	@Param			request			body		dto.DeleteRole			true						"Role deletion data"	example({"role_id":"123e4567-e89b-12d3-a456-426614174000","channel_id":"987fcdeb-51d2-12d3-a456-426614174000"})
//	@Success		200				{object}	map[string]interface{}	"Role ID"					example({"role_id":"123e4567-e89b-12d3-a456-426614174000"})
//	@Failure		400				{object}	map[string]interface{}	"Invalid request"			example({"error":"invalid body"})
//	@Failure		403				{object}	map[string]interface{}	"Insufficient permissions"	example({"error":"the user does not have sufficient rights"})
//	@Failure		500				{object}	map[string]interface{}	"Internal server error"		example({"error":"failed to delete role"})
//	@Router			/roles [delete]
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

	s.Hub.onSendMessage(ctx.Request.Context(), dto.AddMessage{
		ChatID:   body.ChannelID,
		SenderID: uuid.Nil,
		Message:  fmt.Sprintf("User %s deleted the role %s", response.GetName(), roleName),
		IsSystem: true,
	})

	ctx.JSON(http.StatusOK, body.RoleID)
}
