package servicehttp

import (
	"net/http"
	"os"

	"github.com/Flikest/PingVi_backend/internal/delivery/dto"
	"github.com/Flikest/PingVi_backend/pkg/tokens"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

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
