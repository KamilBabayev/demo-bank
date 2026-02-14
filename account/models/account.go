package models

import (
	"time"

	"github.com/shopspring/decimal"
)

// Account types
const (
	AccountTypeChecking = "checking"
	AccountTypeSavings  = "savings"
)

// Account statuses
const (
	AccountStatusActive = "active"
	AccountStatusFrozen = "frozen"
	AccountStatusClosed = "closed"
)

// Savings account withdrawal limit
const SavingsDailyWithdrawalLimit = 5000.00

type Account struct {
	ID                  int64           `json:"id"`
	UserID              int64           `json:"user_id"`
	AccountNumber       string          `json:"account_number"`
	AccountType         string          `json:"account_type"`
	Balance             decimal.Decimal `json:"balance"`
	Currency            string          `json:"currency"`
	Status              string          `json:"status"`
	DailyWithdrawalUsed decimal.Decimal `json:"-"`
	LastWithdrawalDate  *time.Time      `json:"-"`
	CreatedAt           time.Time       `json:"created_at"`
	UpdatedAt           time.Time       `json:"updated_at"`
}

type CreateAccountRequest struct {
	UserID      int64  `json:"user_id" binding:"required"`
	AccountType string `json:"account_type" binding:"required,oneof=checking savings"`
	Currency    string `json:"currency" binding:"omitempty,len=3"`
}

type UpdateAccountRequest struct {
	Status *string `json:"status,omitempty" binding:"omitempty,oneof=active frozen closed"`
}

type DepositRequest struct {
	Amount   decimal.Decimal `json:"amount" binding:"required"`
	Currency string          `json:"currency" binding:"omitempty,len=3"`
}

type WithdrawRequest struct {
	Amount   decimal.Decimal `json:"amount" binding:"required"`
	Currency string          `json:"currency" binding:"omitempty,len=3"`
}

type BalanceResponse struct {
	AccountID     int64           `json:"account_id"`
	AccountNumber string          `json:"account_number"`
	Balance       decimal.Decimal `json:"balance"`
	Currency      string          `json:"currency"`
	AccountType   string          `json:"account_type"`
	Status        string          `json:"status"`
}

type AccountListResponse struct {
	Accounts []Account `json:"accounts"`
	Total    int64     `json:"total"`
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
	Status        string `json:"status"` // "completed" or "failed"
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
	Status        string `json:"status"` // "completed" or "failed"
	FailureReason string `json:"failure_reason,omitempty"`
	AccountID     int64  `json:"account_id,omitempty"`
	UserID        int64  `json:"user_id,omitempty"`
}
