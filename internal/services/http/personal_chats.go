package servicehttp

import (
	"net/http"
	"os"

	"github.com/Flikest/PingVi_backend/internal/delivery/dto"
	"github.com/Flikest/PingVi_backend/pkg/tokens"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// CreatePersonalChat godoc
//
//	@Summary		Create a personal chat
//	@Description	Creates a new direct message chat between two users.
//	@Description	- If chat already exists, returns existing chat
//	@Description	- Both users must be authenticated
//	@Tags			messenger-personal-chat
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			Authorization	header		string					true	"Bearer JWT token"		example("Bearer eyJhbGciOiJIUzI1NiIs...")
//	@Param			request			body		dto.PersonalChat		true	"Personal chat data"	example({"user2_id":"987fcdeb-51d2-12d3-a456-426614174000"})
//	@Success		201				{object}	dto.PersonalChat		"Created personal chat"
//	@Failure		400				{object}	map[string]interface{}	"Invalid request"		example({"error":"invalid body"})
//	@Failure		500				{object}	map[string]interface{}	"Internal server error"	example({"error":"failed to create personal chat"})
//	@Router			/personal_chat [post]
func (s *ServiceMessenger) CreatePersonalChat(ctx *gin.Context) {
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

	var body dto.PersonalChat
	if err := ctx.BindJSON(&body); err != nil {
		s.Log.Error("invalid body: ", "error", err)
		ctx.JSON(http.StatusBadRequest, err)
		return
	}

	body.User1ID = userID

	personalChat, err := s.Repository.CreatePersonalChat(ctx.Request.Context(), body)
	if err != nil {
		s.Log.Error("error with creating personal chat: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusCreated, personalChat)
}

// DeletePersonalChat godoc
//
//	@Summary		Delete a personal chat
//	@Description	Permanently deletes a personal chat. User must be a participant.
//	@Description	- Only participants can delete the chat
//	@Tags			messenger-personal-chat
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			Authorization	header		string					true					"Bearer JWT token"	example("Bearer eyJhbGciOiJIUzI1NiIs...")
//	@Param			request			body		dto.DeletePersonalChat	true					"Personal chat ID"	example({"id":"123e4567-e89b-12d3-a456-426614174000"})
//	@Success		200				{object}	map[string]interface{}	"Chat ID"				example({"chat_id":"123e4567-e89b-12d3-a456-426614174000"})
//	@Failure		400				{object}	map[string]interface{}	"Invalid request"		example({"error":"invalid body"})
//	@Failure		403				{object}	map[string]interface{}	"Not a participant"		example({"error":"not belongs to chat"})
//	@Failure		500				{object}	map[string]interface{}	"Internal server error"	example({"error":"failed to delete personal chat"})
//	@Router			/personal_chat [delete]
func (s *ServiceMessenger) DeletePersonalChat(ctx *gin.Context) {
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

	var body dto.DeletePersonalChat
	if err := ctx.BindJSON(&body); err != nil {
		s.Log.Error("invalid body: ", "error", err)
		ctx.JSON(http.StatusBadRequest, err)
		return
	}

	isBelongs, err := s.Repository.IsBelongsToChat(ctx.Request.Context(), userID, body.ID)
	if err != nil {
		s.Log.Error("error with checking belongs to chat: ", "error", err)
		ctx.JSON(http.StatusBadRequest, err)
		return
	}

	if !isBelongs {
		ctx.JSON(http.StatusForbidden, gin.H{"error": "not belongs to chat"})
		return
	}

	body.UserID = userID

	chatID, err := s.Repository.DeletePersonalChat(ctx.Request.Context(), body)
	if err != nil {
		s.Log.Error("error with deleting personal chat: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, chatID)
}
