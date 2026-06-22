package dto

import (
	"time"

	"github.com/google/uuid"
)

type Topic struct {
	ID          uuid.UUID `json:"id"`
	ChannelId   uuid.UUID `json:"channel_id"`
	DirectoryID string    `json:"directory_id"`
	Name        string    `json:"name"`
	TopicType   string    `json:"topic_type"`
	Position    int       `json:"position"`
	IsPrivate   bool      `json:"is_private"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreateTopic struct {
	UserID    uuid.UUID `json:"user_id"`
	ChannelId uuid.UUID `json:"channel_id"`
	Name      string    `json:"name"`
	TopicType string    `json:"topic_type"`
	Position  int       `json:"position"`
	IsPrivate bool      `json:"is_private"`
}

type UpdateTopic struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	ChannelId uuid.UUID `json:"channel_id"`
	Name      string    `json:"name"`
	TopicType string    `json:"topic_type"`
	Position  int       `json:"position"`
	IsPrivate bool      `json:"is_private"`
}

type DeleteTopic struct {
	ID        uuid.UUID `json:"id"`
	UserId    uuid.UUID `json:"user_id"`
	ChannelId uuid.UUID `json:"channel_id"`
}
