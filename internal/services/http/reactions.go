package servicehttp

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (s *ServiceMessenger) GetAllReactionsFromChat(ctx *gin.Context) {
	chatIDStr := ctx.Param("chat_id")

	chatID, err := uuid.Parse(chatIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid chat id"})
		return
	}

	reactions, err := s.Repository.SelectReactions(ctx.Request.Context(), chatID)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, reactions)
}
