package dto

import (
	"time"

	"github.com/google/uuid"
)

type Bot struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Avatar      []string  `json:"avatar"`
	CreatorID   uuid.UUID `json:"creator_id"`
	Token       string    `json:"token"`
	CreatedAt   time.Time `json:"created_at"`
}

type BotComand struct {
	ID          uuid.UUID `json:"id"`
	BotID       uuid.UUID `json:"bot_id"`
	Command     string    `json:"command"`
	Description string    `json:"description"`
}

type CreateBotRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Avatar      string `json:"avatar"`
}

type CreateBotCommandRequest struct {
	BotID       uuid.UUID `json:"bot_id"`
	Command     string    `json:"command"`
	Description string    `json:"description"`
}

type BotResponse struct {
	Bot      Bot         `json:"bot"`
	Commands []BotComand `json:"commands"`
}
