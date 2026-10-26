package dto

import (
	"time"

	"github.com/gocql/gocql"
)

type Role struct {
	ID            gocql.UUID `json:"id"`
	ChannelID     gocql.UUID `json:"channel_id"`
	Name          string     `json:"name"`
	Color         string     `json:"color"`
	Permissions   string     `json:"permissions"`
	IsMentionable bool       `json:"is_mentionable"`
	IsDefault     bool       `json:"is_default"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}
type CreateRole struct {
	UserID        gocql.UUID `json:"user_id"`
	ChannelID     gocql.UUID `json:"channel_id"`
	Name          string     `json:"name"`
	Color         string     `json:"color"`
	Permissions   string     `json:"permissions"`
	IsMentionable bool       `json:"is_mentionable"`
	IsDefault     bool       `json:"is_default"`
}

type UpdateRole struct {
	UserID        gocql.UUID `json:"user_id"`
	ID            gocql.UUID `json:"id"`
	ChannelID     gocql.UUID `json:"channel_id"`
	Name          string     `json:"name"`
	Color         string     `json:"color"`
	Permissions   string     `json:"permissions"`
	IsMentionable bool       `json:"is_mentionable"`
	IsDefault     bool       `json:"is_default"`
}

type DeleteRole struct {
	UserID    gocql.UUID `json:"user_id"`
	RoleID    gocql.UUID `json:"role_id"`
	ChannelID gocql.UUID `json:"channel_id"`
}
