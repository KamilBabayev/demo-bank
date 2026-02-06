package models

import (
	"encoding/json"
	"time"

	"github.com/shopspring/decimal"
)

// Notification types
const (
	NotificationTypeTransferSent       = "transfer_sent"
	NotificationTypeTransferReceived   = "transfer_received"
	NotificationTypeTransferFailed     = "transfer_failed"
	NotificationTypePaymentProcessed   = "payment_processed"
	NotificationTypePaymentFailed      = "payment_failed"
	NotificationTypeAccountCreated     = "account_created"
	NotificationTypeAccountFrozen      = "account_frozen"
	NotificationTypeLowBalance         = "low_balance"
)

// Notification channels
const (
	ChannelEmail = "email"
	ChannelSMS   = "sms"
	ChannelPush  = "push"
)

// Notification statuses
const (
	NotificationStatusPending = "pending"
	NotificationStatusSent    = "sent"
	NotificationStatusFailed  = "failed"
	NotificationStatusRead    = "read"
)

type Notification struct {
	ID        int64            `json:"id"`
	UserID    int64            `json:"user_id"`
	Type      string           `json:"type"`
	Channel   string           `json:"channel"`
	Title     string           `json:"title"`
	Content   string           `json:"content"`
	Metadata  *json.RawMessage `json:"metadata,omitempty"`
	Status    string           `json:"status"`
	ReadAt    *time.Time       `json:"read_at,omitempty"`
	SentAt    *time.Time       `json:"sent_at,omitempty"`
	CreatedAt time.Time        `json:"created_at"`
	UpdatedAt time.Time        `json:"updated_at"`
}

type CreateNotificationRequest struct {
	UserID   int64            `json:"user_id" binding:"required"`
	Type     string           `json:"type" binding:"required"`
	Channel  string           `json:"channel" binding:"required,oneof=email sms push"`
	Title    string           `json:"title" binding:"required"`
	Content  string           `json:"content" binding:"required"`
	Metadata *json.RawMessage `json:"metadata,omitempty"`
}

type NotificationListResponse struct {
	Notifications []Notification `json:"notifications"`
	Total         int64          `json:"total"`
	Unread        int64          `json:"unread"`
}

// TransferEvent represents a Kafka event for transfers
type TransferEvent struct {
	TransferID    int64           `json:"transfer_id"`
	ReferenceID   string          `json:"reference_id"`
	FromAccountID int64           `json:"from_account_id"`
	ToAccountID   int64           `json:"to_account_id"`
	Amount        decimal.Decimal `json:"amount"`
	Currency      string          `json:"currency"`
}

// TransferResultEvent represents the result of a transfer
type TransferResultEvent struct {
	TransferID    int64  `json:"transfer_id"`
	ReferenceID   string `json:"reference_id"`
	Status        string `json:"status"`
	FailureReason string `json:"failure_reason,omitempty"`
	FromAccountID int64  `json:"from_account_id,omitempty"`
	ToAccountID   int64  `json:"to_account_id,omitempty"`
	FromUserID    int64  `json:"from_user_id,omitempty"`
	ToUserID      int64  `json:"to_user_id,omitempty"`
}

// PaymentEvent represents a Kafka event for payments
type PaymentEvent struct {
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

// PaymentResultEvent represents the result of a payment
type PaymentResultEvent struct {
	PaymentID     int64  `json:"payment_id"`
	ReferenceID   string `json:"reference_id"`
	Status        string `json:"status"`
	FailureReason string `json:"failure_reason,omitempty"`
}
