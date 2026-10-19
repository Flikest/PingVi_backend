package dto

import (
	"time"

	"github.com/google/uuid"
)

type Snapshot struct {
	UserID       uuid.UUID `json:"user_id"`
	AccountType  string    `json:"account_type"`
	BalanceValue float64   `json:"balance_value"`
	At           time.Time `json:"at"`
}
