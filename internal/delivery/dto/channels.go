package dto

import (
	"time"

	"github.com/google/uuid"
)

type Channel struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	OwnerID     uuid.UUID `json:"owner_id"`
	IconURL     string    `json:"icon_url"`
	IsPublic    bool      `json:"is_public"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreateChannelRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	IconURL     string `json:"icon_url"`
	IsPublic    bool   `json:"is_public"`
}

type CreateChannelInput struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	OwnerID     uuid.UUID `json:"owner_id"`
	IconURL     string    `json:"icon_url"`
	IsPublic    bool      `json:"is_public"`
}

type UpdateChannelRequest struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	OwnerID     uuid.UUID `json:"owner_id"`
	IconURL     string    `json:"icon_url"`
	IsPublic    bool      `json:"is_public"`
	CreatedAt   time.Time `json:"created_at"`
}

type UpdateChannelInput struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	OwnerID     uuid.UUID `json:"owner_id"`
	IconURL     string    `json:"icon_url"`
	IsPublic    bool      `json:"is_public"`
	CreatedAt   time.Time `json:"created_at"`
}

type KickChannelMemberRequest struct {
	KickedMemberID uuid.UUID `json:"kicked_user_id"`
	ChannelID      uuid.UUID `json:"channel_id"`
}

type KickChannelMemberInput struct {
	MemberID       uuid.UUID `json:"user_id"`
	KickedMemberID uuid.UUID `json:"kicked_user_id"`
	ChannelID      uuid.UUID `json:"channel_id"`
}

type DeleteChannelRequest struct {
	ID uuid.UUID `json:"id"`
}

type DeleteChannelInput struct {
	ID      uuid.UUID `json:"id"`
	UserID  uuid.UUID `json:"user_id"`
	OwnerID uuid.UUID `json:"owner_id"`
}
