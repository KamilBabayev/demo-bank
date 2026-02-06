package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// Card types
const (
	CardTypeDebit   = "debit"
	CardTypeCredit  = "credit"
	CardTypeVirtual = "virtual"
)

// Card statuses
const (
	CardStatusActive    = "active"
	CardStatusBlocked   = "blocked"
	CardStatusExpired   = "expired"
	CardStatusCancelled = "cancelled"
)

// Card represents a bank card
type Card struct {
	ID                  int64           `json:"id"`
	AccountID           int64           `json:"account_id"`
	CardNumber          string          `json:"card_number"` // Masked: **** **** **** 1234
	CardNumberHash      string          `json:"-"`           // SHA-256 hash for lookups
	CardType            string          `json:"card_type"`
	CardholderName      string          `json:"cardholder_name"`
	ExpirationMonth     int             `json:"expiration_month"`
	ExpirationYear      int             `json:"expiration_year"`
	CVVHash             string          `json:"-"` // Never exposed
	PINHash             string          `json:"-"` // Never exposed
	Status              string          `json:"status"`
	DailyLimit          decimal.Decimal `json:"daily_limit"`
	MonthlyLimit        decimal.Decimal `json:"monthly_limit"`
	PerTransactionLimit decimal.Decimal `json:"per_transaction_limit"`
	DailyUsed           decimal.Decimal `json:"daily_used"`
	MonthlyUsed         decimal.Decimal `json:"monthly_used"`
	LastUsageDate       *time.Time      `json:"-"`
	CreatedAt           time.Time       `json:"created_at"`
	UpdatedAt           time.Time       `json:"updated_at"`
}

// CreateCardRequest represents a request to create a new card
type CreateCardRequest struct {
	AccountID      int64  `json:"account_id" binding:"required"`
	CardType       string `json:"card_type" binding:"required,oneof=debit credit virtual"`
	CardholderName string `json:"cardholder_name" binding:"required"`
}

// UpdateCardRequest represents a request to update card limits or status
type UpdateCardRequest struct {
	Status              *string          `json:"status,omitempty" binding:"omitempty,oneof=active blocked expired cancelled"`
	DailyLimit          *decimal.Decimal `json:"daily_limit,omitempty"`
	MonthlyLimit        *decimal.Decimal `json:"monthly_limit,omitempty"`
	PerTransactionLimit *decimal.Decimal `json:"per_transaction_limit,omitempty"`
}

// SetPINRequest represents a request to set or change the card PIN
type SetPINRequest struct {
	PIN string `json:"pin" binding:"required,len=4"`
}

// CardListResponse represents a paginated list of cards
type CardListResponse struct {
	Cards []Card `json:"cards"`
	Total int64  `json:"total"`
}

// CardEvent represents a Kafka event for card operations
type CardEvent struct {
	CardID         int64  `json:"card_id"`
	AccountID      int64  `json:"account_id"`
	CardType       string `json:"card_type"`
	CardholderName string `json:"cardholder_name"`
	Status         string `json:"status"`
	EventType      string `json:"event_type"` // created, blocked, activated, cancelled
}
