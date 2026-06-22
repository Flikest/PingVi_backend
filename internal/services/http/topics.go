package servicehttp

import (
	"net/http"

	"github.com/Flikest/PingVi_backend/internal/delivery/dto"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (s *ServiceMessenger) CreateTopic(ctx *gin.Context) {
	var body dto.CreateTopic
	if err := ctx.BindJSON(&body); err != nil {
		s.Log.Error("invalid body: ", "error", err)
		ctx.JSON(http.StatusBadRequest, err)
		return
	}

	topic, err := s.Repository.CreateTopic(ctx.Request.Context(), body)
	if err != nil {
		s.Log.Error("error with creating topic: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}

	// TODO добавить расслку о том что пользователь создал топик

	ctx.JSON(http.StatusCreated, topic)
}

func (s *ServiceMessenger) GetAllTopics(ctx *gin.Context) {
	channelID, err := uuid.Parse(ctx.Param("channel_id"))
	if err != nil {
		s.Log.Error("error with parsing user id uuid: ", "error", err)
		ctx.JSON(http.StatusBadRequest, err)
		return
	}

	topics, err := s.Repository.SelectTopicsFromChannel(ctx.Request.Context(), channelID)
	if err != nil {
		s.Log.Error("error with getting topics from channel: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}

	ctx.JSON(http.StatusOK, topics)
}

func (s *ServiceMessenger) UpdateTopic(ctx *gin.Context) {
	var body dto.UpdateTopic
	if err := ctx.BindJSON(&body); err != nil {
		s.Log.Error("invalid body: ", "error", err)
		ctx.JSON(http.StatusBadRequest, err)
		return
	}

	topic, err := s.Repository.UpdateTopic(ctx.Request.Context(), body)
	if err != nil {
		s.Log.Error("error with updating topic: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}

	// TODO добавить рассылку о том что пользователль обновил топик канала

	ctx.JSON(http.StatusOK, topic)
}

func (s *ServiceMessenger) DeleteTopic(ctx *gin.Context) {
	var body dto.DeleteTopic
	if err := ctx.BindJSON(&body); err != nil {
		s.Log.Error("invalid body: ", "error", err)
		ctx.JSON(http.StatusBadRequest, err)
		return
	}

	topicID, err := s.Repository.DeleteTopic(ctx.Request.Context(), body)
	if err != nil {
		s.Log.Error("error with updating topic: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, err)
		return
	}

	// TODO добавить рассылку о том что пользователль удалил топик из канала

	ctx.JSON(http.StatusOK, topicID)
}
