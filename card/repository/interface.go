package repository

import (
	"context"

	"card/models"
)

// CardRepo defines the interface for card data access.
type CardRepo interface {
	Create(ctx context.Context, req *models.CreateCardRequest) (*models.Card, error)
	GetByID(ctx context.Context, id int64) (*models.Card, error)
	ListByAccountID(ctx context.Context, accountID int64, limit, offset int) (*models.CardListResponse, error)
	ListAll(ctx context.Context, limit, offset int) (*models.CardListResponse, error)
	ListByUserAccounts(ctx context.Context, accountIDs []int64, limit, offset int) (*models.CardListResponse, error)
	Update(ctx context.Context, id int64, req *models.UpdateCardRequest) (*models.Card, error)
	Block(ctx context.Context, id int64) (*models.Card, error)
	Unblock(ctx context.Context, id int64) (*models.Card, error)
	Cancel(ctx context.Context, id int64) error
	SetPIN(ctx context.Context, id int64, pin string) error
}
