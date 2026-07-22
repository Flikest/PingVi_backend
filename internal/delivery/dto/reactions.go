package dto

import (
	"time"

	"github.com/bwmarrin/snowflake"
	"github.com/google/uuid"
)

type Reaction struct {
	ID        uuid.UUID    `json:"id"`
	ChatID    uuid.UUID    `json:"chat_id"`
	UserID    uuid.UUID    `json:"user_id"`
	MessageID snowflake.ID `json:"message_id"`
	Reaction  string       `json:"reaction"`
	SendedAt  time.Time    `json:"sended_at"`
}
