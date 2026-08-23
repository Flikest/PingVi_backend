package servicegrpc

import (
	"context"

	"github.com/Flikest/PingVi_backend/internal/delivery/dto"
	"github.com/google/uuid"
)

type BotsRepository interface {
	SelectBotByID(ctx context.Context, botID uuid.UUID) (dto.BotResponse, error)
}

func (s *ServiceBots) GetBotInfo(ctx context.Context, stringBotID string) (dto.BotResponse, error) {
	botID, err := uuid.Parse(stringBotID)
	if err != nil {
		s.Log.Error("error with parsing bot id: ", "error", err)
		return dto.BotResponse{}, err
	}

	bot, err := s.Repository.SelectBotByID(ctx, botID)
	if err != nil {
		s.Log.Error("error with getting bot by id: ", "error", err)
		return dto.BotResponse{}, err
	}

	return bot, nil
}
