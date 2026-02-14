package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"transfer/models"
	"transfer/repository"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	transferByIDTTL    = 3 * time.Minute
	transferByRefTTL   = 3 * time.Minute
	transferAccountTTL = 2 * time.Minute
	transferAllTTL     = 1 * time.Minute
)

// CachedTransferRepository wraps a TransferRepository with Redis caching.
type CachedTransferRepository struct {
	repo  *repository.TransferRepository
	redis *redis.Client
}

// NewCachedTransferRepository creates a new cached transfer repository.
func NewCachedTransferRepository(repo *repository.TransferRepository, redisClient *redis.Client) *CachedTransferRepository {
	return &CachedTransferRepository{
		repo:  repo,
		redis: redisClient,
	}
}

func keyTransferByID(id int64) string {
	return fmt.Sprintf("transfer:id:%d", id)
}

func keyTransferByRef(referenceID uuid.UUID) string {
	return fmt.Sprintf("transfer:ref:%s", referenceID.String())
}

func keyTransferByAccount(accountID int64, limit, offset int) string {
	return fmt.Sprintf("transfer:account:%d:%d:%d", accountID, limit, offset)
}

func keyTransferAll(limit, offset int) string {
	return fmt.Sprintf("transfer:all:%d:%d", limit, offset)
}

// Create delegates to the underlying repo and invalidates list caches.
func (c *CachedTransferRepository) Create(ctx context.Context, req *models.CreateTransferRequest) (*models.Transfer, error) {
	transfer, err := c.repo.Create(ctx, req)
	if err != nil {
		return nil, err
	}
	c.invalidateAccountLists(ctx, transfer.FromAccountID)
	c.invalidateAccountLists(ctx, transfer.ToAccountID)
	c.invalidateGlobalLists(ctx)
	return transfer, nil
}

// GetByID checks cache first, falls back to DB.
func (c *CachedTransferRepository) GetByID(ctx context.Context, id int64) (*models.Transfer, error) {
	key := keyTransferByID(id)

	data, err := c.redis.Get(ctx, key).Bytes()
	if err == nil {
		var transfer models.Transfer
		if json.Unmarshal(data, &transfer) == nil {
			return &transfer, nil
		}
	}

	transfer, err := c.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	c.setCache(ctx, key, transfer, transferByIDTTL)
	return transfer, nil
}

// GetByReferenceID checks cache first, falls back to DB.
func (c *CachedTransferRepository) GetByReferenceID(ctx context.Context, referenceID uuid.UUID) (*models.Transfer, error) {
	key := keyTransferByRef(referenceID)

	data, err := c.redis.Get(ctx, key).Bytes()
	if err == nil {
		var transfer models.Transfer
		if json.Unmarshal(data, &transfer) == nil {
			return &transfer, nil
		}
	}

	transfer, err := c.repo.GetByReferenceID(ctx, referenceID)
	if err != nil {
		return nil, err
	}

	c.setCache(ctx, key, transfer, transferByRefTTL)
	return transfer, nil
}

// ListByAccountID checks cache first, falls back to DB.
func (c *CachedTransferRepository) ListByAccountID(ctx context.Context, accountID int64, limit, offset int) (*models.TransferListResponse, error) {
	key := keyTransferByAccount(accountID, limit, offset)

	data, err := c.redis.Get(ctx, key).Bytes()
	if err == nil {
		var result models.TransferListResponse
		if json.Unmarshal(data, &result) == nil {
			return &result, nil
		}
	}

	result, err := c.repo.ListByAccountID(ctx, accountID, limit, offset)
	if err != nil {
		return nil, err
	}

	c.setCache(ctx, key, result, transferAccountTTL)
	return result, nil
}

// ListByAccountIDs delegates directly (complex key space).
func (c *CachedTransferRepository) ListByAccountIDs(ctx context.Context, accountIDs []int64, limit, offset int) (*models.TransferListResponse, error) {
	return c.repo.ListByAccountIDs(ctx, accountIDs, limit, offset)
}

// ListAll checks cache first, falls back to DB.
func (c *CachedTransferRepository) ListAll(ctx context.Context, limit, offset int) (*models.TransferListResponse, error) {
	key := keyTransferAll(limit, offset)

	data, err := c.redis.Get(ctx, key).Bytes()
	if err == nil {
		var result models.TransferListResponse
		if json.Unmarshal(data, &result) == nil {
			return &result, nil
		}
	}

	result, err := c.repo.ListAll(ctx, limit, offset)
	if err != nil {
		return nil, err
	}

	c.setCache(ctx, key, result, transferAllTTL)
	return result, nil
}

// UpdateStatus delegates to repo and invalidates affected caches.
func (c *CachedTransferRepository) UpdateStatus(ctx context.Context, id int64, status string, failureReason *string) (*models.Transfer, error) {
	transfer, err := c.repo.UpdateStatus(ctx, id, status, failureReason)
	if err != nil {
		return nil, err
	}

	c.invalidateTransfer(ctx, transfer)
	return transfer, nil
}

// MarkAsProcessing marks a transfer as processing and invalidates caches.
func (c *CachedTransferRepository) MarkAsProcessing(ctx context.Context, id int64) (*models.Transfer, error) {
	transfer, err := c.repo.MarkAsProcessing(ctx, id)
	if err != nil {
		return nil, err
	}

	c.invalidateTransfer(ctx, transfer)
	return transfer, nil
}

// MarkAsCompleted marks a transfer as completed and invalidates caches.
func (c *CachedTransferRepository) MarkAsCompleted(ctx context.Context, id int64) (*models.Transfer, error) {
	transfer, err := c.repo.MarkAsCompleted(ctx, id)
	if err != nil {
		return nil, err
	}

	c.invalidateTransfer(ctx, transfer)
	return transfer, nil
}

// MarkAsFailed marks a transfer as failed and invalidates caches.
func (c *CachedTransferRepository) MarkAsFailed(ctx context.Context, id int64, reason string) (*models.Transfer, error) {
	transfer, err := c.repo.MarkAsFailed(ctx, id, reason)
	if err != nil {
		return nil, err
	}

	c.invalidateTransfer(ctx, transfer)
	return transfer, nil
}

// invalidateTransfer invalidates all caches related to a transfer.
func (c *CachedTransferRepository) invalidateTransfer(ctx context.Context, transfer *models.Transfer) {
	c.del(ctx, keyTransferByID(transfer.ID))
	c.del(ctx, keyTransferByRef(transfer.ReferenceID))
	c.invalidateAccountLists(ctx, transfer.FromAccountID)
	c.invalidateAccountLists(ctx, transfer.ToAccountID)
	c.invalidateGlobalLists(ctx)
}

// setCache marshals the value and stores it in Redis. Errors are logged, never returned.
func (c *CachedTransferRepository) setCache(ctx context.Context, key string, value interface{}, ttl time.Duration) {
	data, err := json.Marshal(value)
	if err != nil {
		log.Printf("cache: failed to marshal %s: %v", key, err)
		return
	}
	if err := c.redis.Set(ctx, key, data, ttl).Err(); err != nil {
		log.Printf("cache: failed to set %s: %v", key, err)
	}
}

// del deletes keys from Redis. Errors are logged, never returned.
func (c *CachedTransferRepository) del(ctx context.Context, keys ...string) {
	if err := c.redis.Del(ctx, keys...).Err(); err != nil {
		log.Printf("cache: failed to delete keys: %v", err)
	}
}

// invalidateAccountLists removes cached list entries for an account using a pattern scan.
func (c *CachedTransferRepository) invalidateAccountLists(ctx context.Context, accountID int64) {
	pattern := fmt.Sprintf("transfer:account:%d:*", accountID)
	c.deleteByPattern(ctx, pattern)
}

// invalidateGlobalLists removes all paginated list caches.
func (c *CachedTransferRepository) invalidateGlobalLists(ctx context.Context) {
	c.deleteByPattern(ctx, "transfer:all:*")
}

// deleteByPattern scans and deletes keys matching a pattern.
func (c *CachedTransferRepository) deleteByPattern(ctx context.Context, pattern string) {
	iter := c.redis.Scan(ctx, 0, pattern, 100).Iterator()
	for iter.Next(ctx) {
		c.del(ctx, iter.Val())
	}
	if err := iter.Err(); err != nil {
		log.Printf("cache: failed to scan pattern %s: %v", pattern, err)
	}
}
