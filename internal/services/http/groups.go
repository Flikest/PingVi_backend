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

func (s *ServiceMessenger) CreateGroup(ctx *gin.Context) {
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

	var body dto.CreateGroup
	if err := ctx.BindJSON(&body); err != nil {
		s.Log.Error("invalid body: ", "error", err)
		ctx.JSON(http.StatusBadRequest, err)
		return
	}

	body.OwnerID = userID

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

	response, err := s.Client.GetUserNameByID(ctx.Request.Context(), &pb.GetUserNameByIdRequest{
		UserId: userID.String(),
	})
	if err != nil {
		s.Log.Error("error with getting user name: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	s.Hub.onSendMessage(dto.AddMessage{
		ChatID:      groupID,
		SenderID:    uuid.Nil,
		Message:     fmt.Sprintf("User %s create this group", response.GetName()),
		MessageType: "system",
	})

	ctx.JSON(http.StatusCreated, group)
}

func (s *ServiceMessenger) JoinGroup(ctx *gin.Context) {
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

	groupID, err := uuid.Parse(ctx.Param("group_id"))
	if err != nil {
		s.Log.Error("error with parsing group id uuid", "error", err)
		ctx.JSON(http.StatusBadRequest, err)
		return
	}

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

	response, err := s.Client.GetUserNameByID(ctx.Request.Context(), &pb.GetUserNameByIdRequest{
		UserId: userID.String(),
	})
	if err != nil {
		s.Log.Error("error with getting user name: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	s.Hub.onSendMessage(dto.AddMessage{
		ChatID:      groupID,
		SenderID:    uuid.Nil,
		Message:     fmt.Sprintf("User %s joined the channel", response.GetName()),
		MessageType: "system",
	})

	ctx.JSON(http.StatusOK, groupID)
}

func (s *ServiceMessenger) LeaveGroup(ctx *gin.Context) {
	groupID, err := uuid.Parse(ctx.Param("group_id"))
	if err != nil {
		s.Log.Error("error with parsing group id uuid", "error", err)
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

	ownerID, err := s.Repository.SelectGroupOwnerID(ctx.Request.Context(), groupID)
	if err != nil {
		s.Log.Error("error with checking owner: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}

	if ownerID == userID {
		if err := s.Repository.DeleteGroupMembersByGroupID(ctx.Request.Context(), groupID); err != nil {
			s.Log.Error("error with deleting members", "error", err)
			ctx.JSON(http.StatusInternalServerError, err)
			return
		}

		if err := s.Repository.DeleteGroupByID(ctx.Request.Context(), groupID); err != nil {
			s.Log.Error("error with deleting group", "error", err)
			ctx.JSON(http.StatusInternalServerError, err)
			return
		}

		groupName, err := s.Repository.SelectGroupNameByID(ctx.Request.Context(), groupID)
		if err != nil {
			s.Log.Error("error with getting group name: ", "error", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		s.Hub.onSendMessage(dto.AddMessage{
			ChatID:      groupID,
			SenderID:    uuid.Nil,
			Message:     fmt.Sprintf("channel %s was deleted", groupName),
			MessageType: "system",
		})

		ctx.JSON(http.StatusOK, groupID)
		return
	}

	if err := s.Repository.DeleteGroupMember(ctx.Request.Context(), groupID, userID); err != nil {
		s.Log.Error("error when user leaves group: ", "error", err)
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
		ChatID:      groupID,
		SenderID:    uuid.Nil,
		Message:     fmt.Sprintf("User %s has left the channel", response.GetName()),
		MessageType: "system",
	})

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

	var body dto.CickGroupMember
	if err := ctx.BindJSON(&body); err != nil {
		s.Log.Error("invalid body: ", "error", err)
		ctx.JSON(http.StatusBadRequest, err)
		return
	}

	body.MemberID = userID

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

	groups, err := s.Repository.SelectGroup(ctx.Request.Context(), userID)
	if err != nil {
		s.Log.Error("Error searching for all user groups", "error", err)
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, groups)
}

func (s *ServiceMessenger) UpdateGroup(ctx *gin.Context) {
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

	var body dto.UpdateGroup
	if err := ctx.BindJSON(&body); err != nil {
		s.Log.Error("invalid body: ", "error", err)
		ctx.JSON(http.StatusBadRequest, err)
		return
	}

	body.UserID = userID

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

	response, err := s.Client.GetUserNameByID(ctx.Request.Context(), &pb.GetUserNameByIdRequest{
		UserId: userID.String(),
	})
	if err != nil {
		s.Log.Error("error with getting user name: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	s.Hub.onSendMessage(dto.AddMessage{
		ChatID:      group.ID,
		SenderID:    uuid.Nil,
		Message:     fmt.Sprintf("User %s changed the group", response.GetName()),
		MessageType: "system",
	})

	ctx.JSON(http.StatusOK, group)
}

func (s *ServiceMessenger) DeleteGroup(ctx *gin.Context) {
	var body dto.DeleteGroup
	if err := ctx.BindJSON(&body); err != nil {
		s.Log.Error("invalid body: ", "error", err)
		ctx.JSON(http.StatusBadRequest, err)
		return
	}

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

	if err := s.Repository.DeleteGroupMembersByGroupID(ctx.Request.Context(), body.ID); err != nil {
		s.Log.Error("error with deleting members: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}

	if err := s.Repository.DeleteGroupByID(ctx.Request.Context(), body.ID); err != nil {
		s.Log.Error("error with deleting group", "error", err)
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}

	groupName, err := s.Repository.SelectGroupNameByID(ctx.Request.Context(), body.ID)
	if err != nil {
		s.Log.Error("error with selecting channel name: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	s.Hub.onSendMessage(dto.AddMessage{
		ChatID:      body.ID,
		SenderID:    uuid.Nil,
		Message:     fmt.Sprintf("channel %s was deleted", groupName),
		MessageType: "system",
	})

	ctx.JSON(http.StatusOK, body.ID)
}
