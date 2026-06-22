package dto

import "github.com/google/uuid"

type CreateJoinToken struct {
	UserID  uuid.UUID `json:"user_id"`
	TopicID uuid.UUID `json:"topic_id"`
	Name    string    `json:"name"`
}
