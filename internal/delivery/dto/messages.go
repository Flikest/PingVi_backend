package dto

import (
	"time"

	"github.com/bwmarrin/snowflake"
	"github.com/gocql/gocql"
)

type Message struct {
	ID        snowflake.ID `json:"id"`
	ChatID    gocql.UUID   `json:"chat_id"`
	SenderID  gocql.UUID   `json:"sender_id"`
	Message   string       `json:"message"`
	IsMy      bool         `json:"is_my"`
	IsSystem  bool         `json:"is_system"`
	ReplyToID snowflake.ID `json:"reply_to_id"`
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`
}

type MessageFromClient struct {
	ID        snowflake.ID `json:"id"`
	ChatID    gocql.UUID   `json:"chat_id"`
	Message   string       `json:"message"`
	ReplyToID snowflake.ID `json:"reply_to_id"`
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`
}

type LinkPreviewResponse struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

type ReadMessage struct {
	ChatID    gocql.UUID   `json:"chat_id"`
	UserID    gocql.UUID   `json:"user_id"`
	MessageID snowflake.ID `json:"message_id"`
	At        time.Time    `json:"at"`
}

type AddMessage struct {
	ChatID   gocql.UUID `json:"chat_id"`
	SenderID gocql.UUID `json:"sender_id"`
	Message  string     `json:"message"`
	IsSystem bool       `json:"is_system"`
}

type LeaveMemberMessage struct {
	ChatID  gocql.UUID `json:"chat_id"`
	Message string     `json:"message"`
}

type KickMemberMessage struct {
	UserID  gocql.UUID `json:"user_id"`
	ChatID  gocql.UUID `json:"chat_id"`
	Message string     `json:"message"`
}

type UpdateMessage struct {
	ID        snowflake.ID `json:"id"`
	UserID    gocql.UUID   `json:"user_id"`
	ChatID    gocql.UUID   `json:"chat_id"`
	SenderID  gocql.UUID   `json:"sender_id"`
	Message   string       `json:"message"`
	ReplyToID snowflake.ID `json:"reply_to_id"`
	CreatedAt time.Time    `json:"created_at"`
}

type ClearMessage struct {
	CommunityType string     `json:"community_type"`
	CommunityID   gocql.UUID `json:"chat_id"`
	UserID        gocql.UUID `json:"user_id"`
}

type DeleteMessage struct {
	ID       snowflake.ID `json:"id"`
	ChatID   gocql.UUID   `json:"chat_id"`
	SenderID gocql.UUID   `json:"sender_id"`
}

type DeleteChatMessage struct {
	ChatID gocql.UUID `json:"chat_id"`
}
