package dto

import (
	"time"

	"github.com/gocql/gocql"
)

type PersonalChat struct {
	ID        gocql.UUID `db:"id"         json:"id"`
	User1ID   gocql.UUID `db:"user1_id"   json:"user1_id"`
	User2ID   gocql.UUID `db:"user2_id"   json:"user2_id"`
	CreatedAt time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt time.Time  `db:"updated_at" json:"updated_at"`
}

type PersonalChatResponse struct {
	ID        gocql.UUID `json:"id"`
	UserID    gocql.UUID `json:"user_id"`
	Name      string     `json:"name"`
	Avatar    string     `json:"avatar"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type DeletePersonalChat struct {
	ID     gocql.UUID `json:"id"`
	UserID gocql.UUID `json:"user_id"`
}
