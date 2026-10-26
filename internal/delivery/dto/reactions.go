package dto

import (
	"time"

	"github.com/bwmarrin/snowflake"
	"github.com/gocql/gocql"
)

type Reaction struct {
	ID        gocql.UUID   `json:"id"`
	ChatID    gocql.UUID   `json:"chat_id"`
	UserID    gocql.UUID   `json:"user_id"`
	MessageID snowflake.ID `json:"message_id"`
	Reaction  string       `json:"reaction"`
	SendedAt  time.Time    `json:"sended_at"`
}
