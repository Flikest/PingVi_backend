package dto

import "time"

type StartСommunitiesData struct {
	Type           string      `json:"type"`
	Communities    interface{} `json:"communities"`
	LastMessage    string      `json:"last_message"`
	SendedAt       time.Time   `json:"sended_at"`
	IsBot          bool        `json:"is_bot"`
	IsMuted        bool        `json:"is_muted"`
	UnreadMessages int         `json:"unread_messages"`
}
