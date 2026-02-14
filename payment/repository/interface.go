package repository

import (
	"context"

	"payment/models"

	"github.com/google/uuid"
)

// PaymentRepo defines the interface for payment data access.
type PaymentRepo interface {
	Create(ctx context.Context, userID int64, req *models.CreatePaymentRequest) (*models.Payment, error)
	GetByID(ctx context.Context, id int64) (*models.Payment, error)
	GetByReferenceID(ctx context.Context, referenceID uuid.UUID) (*models.Payment, error)
	ListByUserID(ctx context.Context, userID int64, limit, offset int) (*models.PaymentListResponse, error)
	ListByAccountID(ctx context.Context, accountID int64, limit, offset int) (*models.PaymentListResponse, error)
	ListAll(ctx context.Context, limit, offset int) (*models.PaymentListResponse, error)
	UpdateStatus(ctx context.Context, id int64, status string, failureReason *string) (*models.Payment, error)
	MarkAsProcessing(ctx context.Context, id int64) (*models.Payment, error)
	MarkAsCompleted(ctx context.Context, id int64) (*models.Payment, error)
	MarkAsFailed(ctx context.Context, id int64, reason string) (*models.Payment, error)
}
