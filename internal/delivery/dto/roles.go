package dto

import (
	"time"

	"github.com/google/uuid"
)

type Role struct {
	ID            uuid.UUID `json:"id"`
	ChannelID     uuid.UUID `json:"channel_id"`
	Name          string    `json:"name"`
	Color         string    `json:"color"`
	Permissions   string    `json:"permissions"`
	IsMentionable bool      `json:"is_mentionable"`
	IsDefault     bool      `json:"is_default"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type ChannelMember struct {
	ChannelID uuid.UUID `json:"channel_id"`
	UserID    uuid.UUID `json:"user_id"`
	RoleID    uuid.UUID `json:"role_id"`
	JoinedAt  time.Time `json:"joined_at"`
}

type CreateRole struct {
	UserID        uuid.UUID `json:"user_id"`
	ChannelID     uuid.UUID `json:"channel_id"`
	Name          string    `json:"name"`
	Color         string    `json:"color"`
	Permissions   string    `json:"permissions"`
	IsMentionable bool      `json:"is_mentionable"`
	IsDefault     bool      `json:"is_default"`
}

type UpdateRole struct {
	UserID        uuid.UUID `json:"user_id"`
	ID            uuid.UUID `json:"id"`
	ChannelID     uuid.UUID `json:"channel_id"`
	Name          string    `json:"name"`
	Color         string    `json:"color"`
	Permissions   string    `json:"permissions"`
	IsMentionable bool      `json:"is_mentionable"`
	IsDefault     bool      `json:"is_default"`
}

type DeleteRole struct {
	UserID    uuid.UUID `json:"user_id"`
	RoleID    uuid.UUID `json:"role_id"`
	ChannelID uuid.UUID `json:"channel_id"`
}
