package dto

import (
	"time"

	"github.com/gocql/gocql"
)

type UserStatus struct {
	UserID    gocql.UUID `json:"user_id"`
	Status    string     `json:"status"`
	Timestamp time.Time  `json:"timestamp"`
}

type UserPresence struct {
	UserID     gocql.UUID `json:"user_id"`
	ActiveChat gocql.UUID `json:"active_chat,omitempty"`
	TypingChat gocql.UUID `json:"typing_chat,omitempty"`
	IsOnline   bool       `json:"is_online"`
	LastSeen   time.Time  `json:"last_seen,omitempty"`
}

type TypingStatus struct {
	UserID    gocql.UUID `json:"user_id"`
	ChatID    gocql.UUID `json:"chat_id"`
	IsTyping  bool       `json:"is_typing"`
	Timestamp time.Time  `json:"timestamp"`
}

type StatusRequest struct {
	UserIDs []gocql.UUID `json:"user_ids"`
}

type StatusResponse struct {
	Statuses map[gocql.UUID]UserStatus `json:"statuses"`
}

type PresenceResponse struct {
	Presences map[gocql.UUID]UserPresence `json:"presences"`
}

type ActiveChatUpdate struct {
	UserID gocql.UUID `json:"user_id"`
	ChatID gocql.UUID `json:"chat_id"`
}

type TypingUpdate struct {
	UserID   gocql.UUID `json:"user_id"`
	ChatID   gocql.UUID `json:"chat_id"`
	IsTyping bool       `json:"is_typing"`
}
