package dto

import (
	"time"

	"github.com/gocql/gocql"
)

type Topic struct {
	ID          gocql.UUID `json:"id"`
	ChannelId   gocql.UUID `json:"channel_id"`
	DirectoryID string     `json:"directory_id"`
	Name        string     `json:"name"`
	TopicType   string     `json:"topic_type"`
	Position    int        `json:"position"`
	IsPrivate   bool       `json:"is_private"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type CreateTopic struct {
	UserID    gocql.UUID `json:"user_id"`
	ChannelId gocql.UUID `json:"channel_id"`
	Name      string     `json:"name"`
	TopicType string     `json:"topic_type"`
	Position  int        `json:"position"`
	IsPrivate bool       `json:"is_private"`
}

type UpdateTopic struct {
	ID        gocql.UUID `json:"id"`
	UserID    gocql.UUID `json:"user_id"`
	ChannelId gocql.UUID `json:"channel_id"`
	Name      string     `json:"name"`
	TopicType string     `json:"topic_type"`
	Position  int        `json:"position"`
	IsPrivate bool       `json:"is_private"`
}

type DeleteTopic struct {
	ID        gocql.UUID `json:"id"`
	UserId    gocql.UUID `json:"user_id"`
	ChannelId gocql.UUID `json:"channel_id"`
}
