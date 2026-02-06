package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// Transfer statuses
const (
	TransferStatusPending    = "pending"
	TransferStatusProcessing = "processing"
	TransferStatusCompleted  = "completed"
	TransferStatusFailed     = "failed"
)

type Transfer struct {
	ID            int64           `json:"id"`
	ReferenceID   uuid.UUID       `json:"reference_id"`
	FromAccountID int64           `json:"from_account_id"`
	ToAccountID   int64           `json:"to_account_id"`
	Amount        decimal.Decimal `json:"amount"`
	Currency      string          `json:"currency"`
	Status        string          `json:"status"`
	FailureReason *string         `json:"failure_reason,omitempty"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
	CompletedAt   *time.Time      `json:"completed_at,omitempty"`
}

type CreateTransferRequest struct {
	FromAccountID int64           `json:"from_account_id" binding:"required"`
	ToAccountID   int64           `json:"to_account_id" binding:"required"`
	Amount        decimal.Decimal `json:"amount" binding:"required"`
	Currency      string          `json:"currency" binding:"omitempty,len=3"`
}

type TransferListResponse struct {
	Transfers []Transfer `json:"transfers"`
	Total     int64      `json:"total"`
}

// TransferRequestedEvent is published to Kafka when a transfer is requested
type TransferRequestedEvent struct {
	TransferID    int64           `json:"transfer_id"`
	ReferenceID   string          `json:"reference_id"`
	FromAccountID int64           `json:"from_account_id"`
	ToAccountID   int64           `json:"to_account_id"`
	Amount        decimal.Decimal `json:"amount"`
	Currency      string          `json:"currency"`
}

// TransferResultEvent is consumed from Kafka after account service processes
type TransferResultEvent struct {
	TransferID    int64  `json:"transfer_id"`
	ReferenceID   string `json:"reference_id"`
	Status        string `json:"status"` // "completed" or "failed"
	FailureReason string `json:"failure_reason,omitempty"`
}
