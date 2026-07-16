package dto

import (
	"time"

	"github.com/google/uuid"
)

type UserStatus struct {
	UserID    uuid.UUID `json:"user_id"`
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}

type UserPresence struct {
	UserID     uuid.UUID `json:"user_id"`
	ActiveChat uuid.UUID `json:"active_chat,omitempty"`
	TypingChat uuid.UUID `json:"typing_chat,omitempty"`
	IsOnline   bool      `json:"is_online"`
	LastSeen   time.Time `json:"last_seen,omitempty"`
}

type TypingStatus struct {
	UserID    uuid.UUID `json:"user_id"`
	ChatID    uuid.UUID `json:"chat_id"`
	IsTyping  bool      `json:"is_typing"`
	Timestamp time.Time `json:"timestamp"`
}

type StatusRequest struct {
	UserIDs []uuid.UUID `json:"user_ids"`
}

type StatusResponse struct {
	Statuses map[uuid.UUID]UserStatus `json:"statuses"`
}

type PresenceResponse struct {
	Presences map[uuid.UUID]UserPresence `json:"presences"`
}

type ActiveChatUpdate struct {
	UserID uuid.UUID `json:"user_id"`
	ChatID uuid.UUID `json:"chat_id"`
}

type TypingUpdate struct {
	UserID   uuid.UUID `json:"user_id"`
	ChatID   uuid.UUID `json:"chat_id"`
	IsTyping bool      `json:"is_typing"`
}
