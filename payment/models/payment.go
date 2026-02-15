package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// Payment types
const (
	PaymentTypeBill     = "bill"
	PaymentTypeMerchant = "merchant"
	PaymentTypeExternal = "external"
	PaymentTypeMobile   = "mobile"
)

// MobileOperator represents a mobile operator with its valid prefixes
type MobileOperator struct {
	Name     string   `json:"name"`
	Prefixes []string `json:"prefixes"`
}

// MobileOperators defines valid Azerbaijani mobile operators and their prefixes
var MobileOperators = map[string][]string{
	"Azercell": {"050", "051"},
	"Bakcell":  {"055", "099"},
	"Nar":      {"070", "077"},
}

// Payment statuses
const (
	PaymentStatusPending    = "pending"
	PaymentStatusProcessing = "processing"
	PaymentStatusCompleted  = "completed"
	PaymentStatusFailed     = "failed"
)

type Payment struct {
	ID               int64           `json:"id"`
	ReferenceID      uuid.UUID       `json:"reference_id"`
	AccountID        int64           `json:"account_id"`
	UserID           int64           `json:"user_id"`
	PaymentType      string          `json:"payment_type"`
	RecipientName    *string         `json:"recipient_name,omitempty"`
	RecipientAccount *string         `json:"recipient_account,omitempty"`
	RecipientBank    *string         `json:"recipient_bank,omitempty"`
	Amount           decimal.Decimal `json:"amount"`
	Currency         string          `json:"currency"`
	Description      *string         `json:"description,omitempty"`
	Status           string          `json:"status"`
	FailureReason    *string         `json:"failure_reason,omitempty"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
	ProcessedAt      *time.Time      `json:"processed_at,omitempty"`
}

type CreatePaymentRequest struct {
	AccountID        int64           `json:"account_id" binding:"required"`
	PaymentType      string          `json:"payment_type" binding:"required,oneof=bill merchant external mobile"`
	RecipientName    *string         `json:"recipient_name"`
	RecipientAccount *string         `json:"recipient_account"`
	RecipientBank    *string         `json:"recipient_bank"`
	Amount           decimal.Decimal `json:"amount" binding:"required"`
	Currency         string          `json:"currency" binding:"omitempty,len=3"`
	Description      *string         `json:"description"`
}

type PaymentListResponse struct {
	Payments []Payment `json:"payments"`
	Total    int64     `json:"total"`
}

// PaymentRequestedEvent is published to Kafka when a payment is requested
type PaymentRequestedEvent struct {
	PaymentID        int64           `json:"payment_id"`
	ReferenceID      string          `json:"reference_id"`
	AccountID        int64           `json:"account_id"`
	UserID           int64           `json:"user_id"`
	PaymentType      string          `json:"payment_type"`
	RecipientName    string          `json:"recipient_name,omitempty"`
	RecipientAccount string          `json:"recipient_account,omitempty"`
	Amount           decimal.Decimal `json:"amount"`
	Currency         string          `json:"currency"`
}

// PaymentResultEvent is consumed from Kafka after processing
type PaymentResultEvent struct {
	PaymentID     int64  `json:"payment_id"`
	ReferenceID   string `json:"reference_id"`
	Status        string `json:"status"` // "completed" or "failed"
	FailureReason string `json:"failure_reason,omitempty"`
}
