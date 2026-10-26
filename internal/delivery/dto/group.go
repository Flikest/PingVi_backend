package dto

import (
	"time"

	"github.com/gocql/gocql"
)

type Group struct {
	ID             gocql.UUID `json:"id"`
	Name           string     `json:"name"`
	AvatarLinks    []string   `json:"avatar_links"`
	OwnerID        gocql.UUID `json:"owner_id"`
	InvitationLink string     `json:"invitation_link"`
	IsPublic       bool       `json:"is_publick"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type CreateGroup struct {
	Name        string     `json:"name"`
	AvatarLinks []string   `json:"avatar_links"`
	OwnerID     gocql.UUID `json:"owner_id"`
	IsPublic    bool       `json:"is_publick"`
}

type JoinGroup struct {
	InvitationLink string `json:"invitation_link"` // https://домен/group/join/group_id/invited_user_id
}

type KickGroupMember struct {
	MemberID       gocql.UUID `json:"user_id"`
	KickedMemberID gocql.UUID `json:"kicked_user_id"`
	GroupID        gocql.UUID `json:"channel_id"`
}

type GroupMember struct {
	GroupID  gocql.UUID `json:"group_id"`
	UserID   gocql.UUID `json:"user_id"`
	Name     string     `json:"name"`
	Avatar   string     `json:"avatar"`
	IsMuted  bool       `json:"is_muted"`
	IsAdmin  bool       `json:"is_admin"`
	JoinedAt time.Time  `json:"joined_at"`
}

type UpdateGroup struct {
	UserID      gocql.UUID `json:"user_id"`
	ID          gocql.UUID `json:"id"`
	Name        string     `json:"name"`
	AvatarLinks []string   `json:"avatar_links"`
	OwnerID     gocql.UUID `json:"owner_id"`
	IsPublic    bool       `json:"is_publick"`
	CreatedAt   time.Time  `json:"created_at"`
}

type DeleteGroup struct {
	UserID gocql.UUID `json:"user_id"`
	ID     gocql.UUID `json:"id"`
}
