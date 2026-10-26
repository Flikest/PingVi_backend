package dto

import (
	"time"

	"github.com/gocql/gocql"
)

type Channel struct {
	ID          gocql.UUID `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	OwnerID     gocql.UUID `json:"owner_id"`
	IconURL     string     `json:"icon_url"`
	IsPublic    bool       `json:"is_public"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type CreateChannelRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	IconURL     string `json:"icon_url"`
	IsPublic    bool   `json:"is_public"`
}

type ChannelMember struct {
	ChannelID gocql.UUID `json:"channel_id"`
	UserID    gocql.UUID `json:"user_id"`
	IsBot     bool       `json:"is_bot"`
	IsMuted   bool       `json:"is_muted"`
	RoleID    gocql.UUID `json:"role_id"`
	Nickname  string     `json:"nickname"`
	Avatar    string     `json:"avatar"`
	JoinedAt  time.Time  `json:"joined_at"`
}

type CreateChannelInput struct {
	ID          gocql.UUID `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	OwnerID     gocql.UUID `json:"owner_id"`
	IconURL     string     `json:"icon_url"`
	IsPublic    bool       `json:"is_public"`
}

type UpdateChannelRequest struct {
	ID          gocql.UUID `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	OwnerID     gocql.UUID `json:"owner_id"`
	IconURL     string     `json:"icon_url"`
	IsPublic    bool       `json:"is_public"`
	CreatedAt   time.Time  `json:"created_at"`
}

type UpdateChannelInput struct {
	ID          gocql.UUID `json:"id"`
	UserID      gocql.UUID `json:"user_id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	OwnerID     gocql.UUID `json:"owner_id"`
	IconURL     string     `json:"icon_url"`
	IsPublic    bool       `json:"is_public"`
	CreatedAt   time.Time  `json:"created_at"`
}

type KickChannelMemberRequest struct {
	KickedMemberID gocql.UUID `json:"kicked_user_id"`
	ChannelID      gocql.UUID `json:"channel_id"`
}

type KickChannelMemberInput struct {
	MemberID       gocql.UUID `json:"user_id"`
	KickedMemberID gocql.UUID `json:"kicked_user_id"`
	ChannelID      gocql.UUID `json:"channel_id"`
}

type DeleteChannelRequest struct {
	ID gocql.UUID `json:"id"`
}

type DeleteChannelInput struct {
	ID      gocql.UUID `json:"id"`
	UserID  gocql.UUID `json:"user_id"`
	OwnerID gocql.UUID `json:"owner_id"`
}
