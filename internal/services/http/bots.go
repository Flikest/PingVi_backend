package servicehttp

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/Flikest/PingVi_backend/internal/delivery/dto"
	randomstring "github.com/Flikest/PingVi_backend/pkg/random_string"
	"github.com/Flikest/PingVi_backend/pkg/tokens"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type BotHub struct {
	mu       sync.RWMutex
	Messages map[string]chan []Message
	Answers  chan []Message
}

func (s *ServiceBots) GetMessages(ctx *gin.Context) {
	botToken := ctx.Param("token")
	if botToken == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Bot token missing"})
		return
	}

	s.BotHub.mu.Lock()
	if s.BotHub.Messages == nil {
		s.BotHub.Messages = make(map[string]chan []Message)
	}

	botID, _, _ := strings.Cut(botToken, ":")

	messagesChan, ok := s.BotHub.Messages[botID]
	if !ok {
		messagesChan = make(chan []Message, 100)
		s.BotHub.Messages[botToken] = messagesChan
	}
	s.BotHub.mu.Unlock()

	timeOutContext, cancel := context.WithTimeout(ctx.Request.Context(), 1*time.Minute)
	defer cancel()

	select {
	case data := <-messagesChan:
		ctx.JSON(http.StatusOK, data)
	case <-timeOutContext.Done():
		ctx.Status(http.StatusNoContent)
	}
}

func (s *ServiceBots) SendAnswers(ctx *gin.Context) {
	botToken := ctx.Param("token")
	if botToken == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Bot token missing"})
		return
	}

	var body []Message
	if err := ctx.ShouldBindJSON(&body); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	left, _, _ := strings.Cut(botToken, ":")

	botID, err := uuid.Parse(left)
	if err != nil {
		s.Log.Error("error with parsing uuid bot id: ", "error", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var validatedMessages []Message

	for _, j := range body {
		j.Message.SenderID = botID
		j.Reaction.UserID = botID
		validatedMessages = append(validatedMessages, j)
	}

	s.BotHub.Answers <- validatedMessages

	ctx.JSON(http.StatusOK, gin.H{"message": "answers delivered"})
}

func (s *ServiceBots) CreateBot(ctx *gin.Context) {
	payload, err := tokens.Verify(ctx.Request.Header.Get("Authorization"), []byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		s.Log.Error("error with verify jwt user token: ", "error", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	creatorID, err := uuid.Parse(payload.ID.String())
	if err != nil {
		s.Log.Error("error with parsing jwt token from header: ", "error", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var body dto.CreateBotRequest

	if err := ctx.ShouldBindJSON(&body); err != nil {
		s.Log.Error("error with binding body: ", "error", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	botID, err := uuid.NewV7()
	if err != nil {
		s.Log.Error("error with generate bot id: ", "error", err)
	}

	str, err := randomstring.GenerateRandomString(32)
	if err != nil {
		s.Log.Error("error with generate random string: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var bot = dto.Bot{
		ID:          botID,
		Name:        body.Name,
		Description: body.Description,
		Avatar:      []string{body.Avatar},
		CreatorID:   creatorID,
		Token:       fmt.Sprintf("%s:%s", botID, str),
		CreatedAt:   time.Now(),
	}

	if err := s.Repository.InsertBot(ctx.Request.Context(), bot); err != nil {
		s.Log.Error("error with creating bot: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, bot)
}

func (s *ServiceBots) GetBotById(ctx *gin.Context) {
	paramBotId := ctx.Param("bot_id")

	botID, err := uuid.Parse(paramBotId)
	if err != nil {
		s.Log.Error("error with parse bot id: ", "error", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	bot, err := s.Repository.SelectBotByID(ctx.Request.Context(), botID)
	if err != nil {
		s.Log.Error("error with getting bot by id: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, bot)
}

func (s *ServiceBots) GetMyBots(ctx *gin.Context) {
	payload, err := tokens.Verify(ctx.Request.Header.Get("Authorization"), []byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		s.Log.Error("error with verify jwt user token: ", "error", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	creatorID, err := uuid.Parse(payload.ID.String())
	if err != nil {
		s.Log.Error("error with parsing jwt token from header: ", "error", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	bots, err := s.Repository.SelectBotsByCreatorID(ctx.Request.Context(), creatorID)
	if err != nil {
		s.Log.Error("error with getting my bots: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, bots)
}

func (s *ServiceBots) UpdateBot(ctx *gin.Context) {
	payload, err := tokens.Verify(ctx.Request.Header.Get("Authorization"), []byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		s.Log.Error("error with verify jwt user token: ", "error", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	creatorID, err := uuid.Parse(payload.ID.String())
	if err != nil {
		s.Log.Error("error with parsing jwt token from header: ", "error", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var body dto.Bot

	if err := ctx.ShouldBindJSON(&body); err != nil {
		s.Log.Error("error with binding body: ", "error", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}

	if creatorID != body.CreatorID {
		s.Log.Warn("The user attempted to update the bot's data without being its creator")
		ctx.JSON(http.StatusForbidden, gin.H{"error": "You are not the creator of the bot"})
		return
	}

	if err := s.Repository.UpdateBot(ctx.Request.Context(), body); err != nil {
		s.Log.Error("error with updating bot: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, body)
}

func (s *ServiceBots) DeleteBot(ctx *gin.Context) {
	payload, err := tokens.Verify(ctx.Request.Header.Get("Authorization"), []byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		s.Log.Error("error with verify jwt user token: ", "error", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	creatorID, err := uuid.Parse(payload.ID.String())
	if err != nil {
		s.Log.Error("error with parsing jwt token from header: ", "error", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	paramBotId := ctx.Param("bot_id")

	if paramBotId == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "bot_id is a required field"})
		return
	}

	botID, err := uuid.Parse(paramBotId)
	if err != nil {
		s.Log.Error("error with paring bot id: ", "error", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	bot, err := s.Repository.SelectBotByID(ctx.Request.Context(), botID)
	if err != nil {
		s.Log.Error("error with getting bot by id: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if bot.Bot.CreatorID != creatorID {
		s.Log.Warn("The user attempted to update the bot's data without being its creator")
		ctx.JSON(http.StatusForbidden, gin.H{"error": "You are not the creator of the bot"})
		return
	}

	if err := s.Repository.DeleteBot(ctx.Request.Context(), botID); err != nil {
		s.Log.Error("error with deleting bot: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, bot)
}
