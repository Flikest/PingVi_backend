package dto

import (
	"time"

	"github.com/gocql/gocql"
)

type Directory struct {
	ID        gocql.UUID `json:"id"`
	ChannelID gocql.UUID `json:"channel_id"`
	Name      string     `json:"name"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type CreateDirectory struct {
	UserID    gocql.UUID `json:"user_id"`
	ID        gocql.UUID `json:"id"`
	ChannelID gocql.UUID `json:"channel_id"`
	Name      string     `json:"name"`
}

type UpdateDirectory struct {
	UserID    gocql.UUID `json:"user_id"`
	ID        gocql.UUID `json:"id"`
	ChannelID gocql.UUID `json:"channel_id"`
	Name      string     `json:"name"`
}

type DeleteDirectory struct {
	UserID    gocql.UUID `json:"user_id"`
	ChannelID gocql.UUID `json:"channel_id"`
	ID        gocql.UUID `json:"id"`
}
