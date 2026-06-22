package servicehttp

import (
	"net/http"

	"github.com/Flikest/PingVi_backend/internal/delivery/dto"
	"github.com/gin-gonic/gin"
)

func (s *ServiceMessenger) CreatePersonalChat(ctx *gin.Context) {
	var body dto.PersonalChat
	if err := ctx.BindJSON(&body); err != nil {
		s.Log.Error("invalid body: ", "error", err)
		ctx.JSON(http.StatusBadRequest, err)
		return
	}

	personalChat, err := s.Repository.CreatePersonalChat(ctx.Request.Context(), body)
	if err != nil {
		s.Log.Error("error with creating personal chat: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusCreated, personalChat)
}

func (s *ServiceMessenger) DeletePersonalChat(ctx *gin.Context) {
	var body dto.DeletePersonalChat
	if err := ctx.BindJSON(&body); err != nil {
		s.Log.Error("invalid body: ", "error", err)
		ctx.JSON(http.StatusBadRequest, err)
		return
	}

	chatID, err := s.Repository.DeletePersonalChat(ctx.Request.Context(), body)
	if err != nil {
		s.Log.Error("error with deleting personal chat: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, chatID)
}
