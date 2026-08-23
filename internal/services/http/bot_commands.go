package servicehttp

import (
	"net/http"
	"os"

	"github.com/Flikest/PingVi_backend/internal/delivery/dto"
	"github.com/Flikest/PingVi_backend/pkg/tokens"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (s *ServiceBots) CreateBotCommand(ctx *gin.Context) {
	var body dto.CreateBotCommandRequest
	if err := ctx.ShouldBindJSON(&body); err != nil {
		s.Log.Error("error with binding json: ", "error", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	bot, err := s.Repository.SelectBotByID(ctx.Request.Context(), body.BotID)
	if err != nil {
		s.Log.Error("error with getting bot by id: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

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

	if bot.Bot.CreatorID != creatorID {
		s.Log.Warn("The user attempted to update the bot's data without being its creator")
		ctx.JSON(http.StatusForbidden, gin.H{"error": "You are not the creator of the bot"})
		return
	}

	for _, j := range bot.Commands {
		if j.Command == body.Command {
			s.Log.Warn("The user attempted to create a command with the same name as one already in the database")
			ctx.JSON(http.StatusConflict, gin.H{"error": "That command already exists"})
			return
		}
	}

	commandID, err := uuid.NewV7()
	if err != nil {
		s.Log.Error("error with generate bot id: ", "error", err)
	}

	command := dto.BotComand{
		ID:          commandID,
		BotID:       body.BotID,
		Command:     body.Command,
		Description: body.Description,
	}

	if err := s.Repository.InsertBotCommand(ctx.Request.Context(), command); err != nil {
		s.Log.Error("error with creating bot command: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, command)
}

func (s *ServiceBots) GetBotCommands(ctx *gin.Context) {
	paramBotID := ctx.Param("bot_id")

	if paramBotID == "" {
		s.Log.Error("The bot ID was not filled in")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "bot_id is a required query param"})
		return
	}

	botId, err := uuid.Parse(paramBotID)
	if err != nil {
		s.Log.Error("error with parsing bot id: ", "error", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	commands, err := s.Repository.SelectBotCommands(ctx.Request.Context(), botId)
	if err != nil {
		s.Log.Error("error with getting bot commands by id: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, commands)
}

func (s *ServiceBots) UpdateBotCommand(ctx *gin.Context) {
	var body dto.BotComand
	if err := ctx.ShouldBindJSON(&body); err != nil {
		s.Log.Error("error with binding json: ", "error", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	bot, err := s.Repository.SelectBotByID(ctx.Request.Context(), body.BotID)
	if err != nil {
		s.Log.Error("error with getting bot by id: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	payload, err := tokens.Verify(ctx.Request.Header.Get("Authorization"), []byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		s.Log.Error("error with verify jwt user token: ", "error", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if bot.Bot.CreatorID != payload.ID {
		s.Log.Warn("The user attempted to update the bot's data without being its creator")
		ctx.JSON(http.StatusForbidden, gin.H{"error": "You are not the creator of the bot"})
		return
	}

	for _, j := range bot.Commands {
		if j.Command == body.Command {
			s.Log.Warn("The user attempted to create a command with the same name as one already in the database")
			ctx.JSON(http.StatusConflict, gin.H{"error": "That command already exists"})
			return
		}
	}

	if err := s.Repository.UpdateBotCommand(ctx.Request.Context(), body); err != nil {
		s.Log.Error("error with updating bot command: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, body)
}

func (s *ServiceBots) DeleteBotCommand(ctx *gin.Context) {
	paramBotID := ctx.Param("bot_id")
	paramCommandID := ctx.Param("command_id")

	if paramCommandID == "" {
		s.Log.Error("The bot ID was not filled in")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "bot_id is a required query param"})
		return
	}
	if paramBotID == "" {
		s.Log.Error("The command ID was not filled in")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "bot_id is a required query param"})
		return
	}

	commandID, err := uuid.Parse(paramCommandID)
	if err != nil {
		s.Log.Error("error with parsing param command_id")
	}

	botID, err := uuid.Parse(paramBotID)
	if err != nil {
		s.Log.Error("error with parsing bot_id")
	}

	bot, err := s.Repository.SelectBotByID(ctx.Request.Context(), botID)
	if err != nil {
		s.Log.Error("error with getting bot by id: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	payload, err := tokens.Verify(ctx.Request.Header.Get("Authorization"), []byte(os.Getenv("JWT_SECRET")))
	if err != nil {
		s.Log.Error("error with verify jwt user token: ", "error", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if bot.Bot.CreatorID != payload.ID {
		s.Log.Warn("The user attempted to update the bot's data without being its creator")
		ctx.JSON(http.StatusForbidden, gin.H{"error": "You are not the creator of the bot"})
		return
	}

	if err := s.Repository.DeleteBotCommand(ctx.Request.Context(), commandID); err != nil {
		s.Log.Error("error with deleting bot command: ", "error", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, commandID)
}
