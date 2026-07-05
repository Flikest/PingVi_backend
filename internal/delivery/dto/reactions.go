package dto

import (
	"time"

	"github.com/google/uuid"
)

type Reaction struct {
	ChatID    uuid.UUID `json:"chat_id"`
	UserID    uuid.UUID `json:"user_id"`
	MessageID uuid.UUID `json:"message_id"`
	Reaction  string    `json:"reaction"`
	SendedAt  time.Time `json:"sended_at"`
}
