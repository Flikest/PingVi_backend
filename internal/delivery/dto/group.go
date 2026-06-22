package dto

import (
	"time"

	"github.com/google/uuid"
)

type Group struct {
	ID             uuid.UUID `json:"id"`
	Name           string    `json:"name"`
	AvatarLinks    []string  `json:"avatar_links"`
	OwnerID        uuid.UUID `json:"owner_id"`
	InvitationLink string    `json:"invitation_link"`
	IsPublic       bool      `json:"is_publick"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type CreateGroup struct {
	Name        string    `json:"name"`
	AvatarLinks []string  `json:"avatar_links"`
	OwnerID     uuid.UUID `json:"owner_id"`
	IsPublic    bool      `json:"is_publick"`
}

type JoinGroup struct {
	InvitationLink string `json:"invitation_link"` // https://домен/group/join/group_id/invited_user_id
}

type CickGroupMember struct {
	MemberID       uuid.UUID `json:"user_id"`
	KickedMemberID uuid.UUID `json:"kicked_user_id"`
	GroupID        uuid.UUID `json:"channel_id"`
}

type GroupMember struct {
	GroupID  uuid.UUID `json:"group_id"`
	UserID   uuid.UUID `json:"user_id"`
	IsMuted  bool      `json:"is_muted"`
	IsAdmin  bool      `json:"is_admin"`
	JoinedAt time.Time `json:"joined_at"`
}

type UpdateGroup struct {
	UserID      uuid.UUID `json:"user_id"`
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	AvatarLinks []string  `json:"avatar_links"`
	OwnerID     uuid.UUID `json:"owner_id"`
	IsPublic    bool      `json:"is_publick"`
	CreatedAt   time.Time `json:"created_at"`
}

type DeleteGroup struct {
	UserID  uuid.UUID `json:"user_id"`
	ID      uuid.UUID `json:"id"`
	OwnerID uuid.UUID `json:"owner_id"`
}
