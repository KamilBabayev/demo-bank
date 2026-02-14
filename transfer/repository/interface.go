package repository

import (
	"context"

	"transfer/models"

	"github.com/google/uuid"
)

// TransferRepo defines the interface for transfer data access.
type TransferRepo interface {
	Create(ctx context.Context, req *models.CreateTransferRequest) (*models.Transfer, error)
	GetByID(ctx context.Context, id int64) (*models.Transfer, error)
	GetByReferenceID(ctx context.Context, referenceID uuid.UUID) (*models.Transfer, error)
	ListByAccountID(ctx context.Context, accountID int64, limit, offset int) (*models.TransferListResponse, error)
	ListByAccountIDs(ctx context.Context, accountIDs []int64, limit, offset int) (*models.TransferListResponse, error)
	ListAll(ctx context.Context, limit, offset int) (*models.TransferListResponse, error)
	UpdateStatus(ctx context.Context, id int64, status string, failureReason *string) (*models.Transfer, error)
	MarkAsProcessing(ctx context.Context, id int64) (*models.Transfer, error)
	MarkAsCompleted(ctx context.Context, id int64) (*models.Transfer, error)
	MarkAsFailed(ctx context.Context, id int64, reason string) (*models.Transfer, error)
}
