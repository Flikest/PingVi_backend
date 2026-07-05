package dto

import (
	"time"

	"github.com/google/uuid"
)

type Message struct {
	CommunityType string    `json:"community_type"`
	ID            uuid.UUID `json:"id"`
	ChatID        uuid.UUID `json:"chat_id"`
	SenderID      uuid.UUID `json:"sender_id"`
	Message       string    `json:"message"`
	MessageType   string    `json:"message_type"`
	IsEdited      bool      `json:"is_edited"`
	ReplyToID     uuid.UUID `json:"reply_to_id"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type LinkPreviewResponse struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

type ReadMessage struct {
	ChatID    uuid.UUID `json:"chat_id"`
	UserID    uuid.UUID `json:"user_id"`
	MessageID uuid.UUID `json:"message_id"`
	At        time.Time `json:"at"`
}

type AddMessage struct {
	ChatID      uuid.UUID `json:"chat_id"`
	SenderID    uuid.UUID `json:"sender_id"`
	Message     string    `json:"message"`
	MessageType string    `json:"message_type"`
	Attachments string    `json:"attachments"`
}

type UpdateMessage struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	ChatID      uuid.UUID `json:"chat_id"`
	SenderID    uuid.UUID `json:"sender_id"`
	Message     string    `json:"message"`
	MessageType string    `json:"message_type"`
	ReplyToID   uuid.UUID `json:"reply_to_id"`
	Attachments string    `json:"attachments"`
	Reactions   []string  `json:"reactions"`
	CreatedAt   time.Time `json:"created_at"`
}

type ClearMessage struct {
	CommunityType string    `json:"community_type"`
	CommunityID   uuid.UUID `json:"chat_id"`
	UserID        uuid.UUID `json:"user_id"`
}

type DeleteMessage struct {
	ID       uuid.UUID `json:"id"`
	ChatID   uuid.UUID `json:"chat_id"`
	SenderID uuid.UUID `json:"sender_id"`
}
