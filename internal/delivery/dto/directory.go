package dto

import (
	"time"

	"github.com/google/uuid"
)

type Directory struct {
	ID        uuid.UUID `json:"id"`
	ChannelID uuid.UUID `json:"channel_id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateDirectory struct {
	UserID    uuid.UUID `json:"user_id"`
	ID        uuid.UUID `json:"id"`
	ChannelID uuid.UUID `json:"channel_id"`
	Name      string    `json:"name"`
}

type UpdateDirectory struct {
	UserID    uuid.UUID `json:"user_id"`
	ID        uuid.UUID `json:"id"`
	ChannelID uuid.UUID `json:"channel_id"`
	Name      string    `json:"name"`
}

type DeleteDirectory struct {
	UserID    uuid.UUID `json:"user_id"`
	channelID uuid.UUID `json:"channel_id"`
	ID        uuid.UUID `json:"id"`
}
