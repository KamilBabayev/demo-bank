package repository

import (
	"context"

	"account/models"

	"github.com/shopspring/decimal"
)

// AccountRepo defines the interface for account data access.
type AccountRepo interface {
	Create(ctx context.Context, req *models.CreateAccountRequest) (*models.Account, error)
	GetByID(ctx context.Context, id int64) (*models.Account, error)
	GetByAccountNumber(ctx context.Context, accountNumber string) (*models.Account, error)
	ListByUserID(ctx context.Context, userID int64, limit, offset int) (*models.AccountListResponse, error)
	ListAllActive(ctx context.Context) (*models.AccountListResponse, error)
	ListAll(ctx context.Context, limit, offset int) (*models.AccountListResponse, error)
	Update(ctx context.Context, id int64, req *models.UpdateAccountRequest) (*models.Account, error)
	Delete(ctx context.Context, id int64) error
	Deposit(ctx context.Context, id int64, amount decimal.Decimal) (*models.Account, error)
	Withdraw(ctx context.Context, id int64, amount decimal.Decimal) (*models.Account, error)
	Transfer(ctx context.Context, fromID, toID int64, amount decimal.Decimal) error
}
