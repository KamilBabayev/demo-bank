package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"payment/models"
	"payment/repository"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	paymentByIDTTL    = 3 * time.Minute
	paymentByRefTTL   = 3 * time.Minute
	paymentUserTTL    = 2 * time.Minute
	paymentAccountTTL = 2 * time.Minute
	paymentAllTTL     = 1 * time.Minute
)

// CachedPaymentRepository wraps a PaymentRepository with Redis caching.
type CachedPaymentRepository struct {
	repo  *repository.PaymentRepository
	redis *redis.Client
}

// NewCachedPaymentRepository creates a new cached payment repository.
func NewCachedPaymentRepository(repo *repository.PaymentRepository, redisClient *redis.Client) *CachedPaymentRepository {
	return &CachedPaymentRepository{
		repo:  repo,
		redis: redisClient,
	}
}

func keyPaymentByID(id int64) string {
	return fmt.Sprintf("payment:id:%d", id)
}

func keyPaymentByRef(referenceID uuid.UUID) string {
	return fmt.Sprintf("payment:ref:%s", referenceID.String())
}

func keyPaymentByUser(userID int64, limit, offset int) string {
	return fmt.Sprintf("payment:user:%d:%d:%d", userID, limit, offset)
}

func keyPaymentByAccount(accountID int64, limit, offset int) string {
	return fmt.Sprintf("payment:account:%d:%d:%d", accountID, limit, offset)
}

func keyPaymentAll(limit, offset int) string {
	return fmt.Sprintf("payment:all:%d:%d", limit, offset)
}

// Create delegates to the underlying repo and invalidates list caches.
func (c *CachedPaymentRepository) Create(ctx context.Context, userID int64, req *models.CreatePaymentRequest) (*models.Payment, error) {
	payment, err := c.repo.Create(ctx, userID, req)
	if err != nil {
		return nil, err
	}
	c.invalidateUserLists(ctx, userID)
	c.invalidateAccountLists(ctx, payment.AccountID)
	c.invalidateGlobalLists(ctx)
	return payment, nil
}

// GetByID checks cache first, falls back to DB.
func (c *CachedPaymentRepository) GetByID(ctx context.Context, id int64) (*models.Payment, error) {
	key := keyPaymentByID(id)

	data, err := c.redis.Get(ctx, key).Bytes()
	if err == nil {
		var payment models.Payment
		if json.Unmarshal(data, &payment) == nil {
			return &payment, nil
		}
	}

	payment, err := c.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	c.setCache(ctx, key, payment, paymentByIDTTL)
	return payment, nil
}

// GetByReferenceID checks cache first, falls back to DB.
func (c *CachedPaymentRepository) GetByReferenceID(ctx context.Context, referenceID uuid.UUID) (*models.Payment, error) {
	key := keyPaymentByRef(referenceID)

	data, err := c.redis.Get(ctx, key).Bytes()
	if err == nil {
		var payment models.Payment
		if json.Unmarshal(data, &payment) == nil {
			return &payment, nil
		}
	}

	payment, err := c.repo.GetByReferenceID(ctx, referenceID)
	if err != nil {
		return nil, err
	}

	c.setCache(ctx, key, payment, paymentByRefTTL)
	return payment, nil
}

// ListByUserID checks cache first, falls back to DB.
func (c *CachedPaymentRepository) ListByUserID(ctx context.Context, userID int64, limit, offset int) (*models.PaymentListResponse, error) {
	key := keyPaymentByUser(userID, limit, offset)

	data, err := c.redis.Get(ctx, key).Bytes()
	if err == nil {
		var result models.PaymentListResponse
		if json.Unmarshal(data, &result) == nil {
			return &result, nil
		}
	}

	result, err := c.repo.ListByUserID(ctx, userID, limit, offset)
	if err != nil {
		return nil, err
	}

	c.setCache(ctx, key, result, paymentUserTTL)
	return result, nil
}

// ListByAccountID checks cache first, falls back to DB.
func (c *CachedPaymentRepository) ListByAccountID(ctx context.Context, accountID int64, limit, offset int) (*models.PaymentListResponse, error) {
	key := keyPaymentByAccount(accountID, limit, offset)

	data, err := c.redis.Get(ctx, key).Bytes()
	if err == nil {
		var result models.PaymentListResponse
		if json.Unmarshal(data, &result) == nil {
			return &result, nil
		}
	}

	result, err := c.repo.ListByAccountID(ctx, accountID, limit, offset)
	if err != nil {
		return nil, err
	}

	c.setCache(ctx, key, result, paymentAccountTTL)
	return result, nil
}

// ListAll checks cache first, falls back to DB.
func (c *CachedPaymentRepository) ListAll(ctx context.Context, limit, offset int) (*models.PaymentListResponse, error) {
	key := keyPaymentAll(limit, offset)

	data, err := c.redis.Get(ctx, key).Bytes()
	if err == nil {
		var result models.PaymentListResponse
		if json.Unmarshal(data, &result) == nil {
			return &result, nil
		}
	}

	result, err := c.repo.ListAll(ctx, limit, offset)
	if err != nil {
		return nil, err
	}

	c.setCache(ctx, key, result, paymentAllTTL)
	return result, nil
}

// UpdateStatus delegates to repo and invalidates affected caches.
func (c *CachedPaymentRepository) UpdateStatus(ctx context.Context, id int64, status string, failureReason *string) (*models.Payment, error) {
	payment, err := c.repo.UpdateStatus(ctx, id, status, failureReason)
	if err != nil {
		return nil, err
	}

	c.invalidatePayment(ctx, payment)
	return payment, nil
}

// MarkAsProcessing marks a payment as processing and invalidates caches.
func (c *CachedPaymentRepository) MarkAsProcessing(ctx context.Context, id int64) (*models.Payment, error) {
	payment, err := c.repo.MarkAsProcessing(ctx, id)
	if err != nil {
		return nil, err
	}

	c.invalidatePayment(ctx, payment)
	return payment, nil
}

// MarkAsCompleted marks a payment as completed and invalidates caches.
func (c *CachedPaymentRepository) MarkAsCompleted(ctx context.Context, id int64) (*models.Payment, error) {
	payment, err := c.repo.MarkAsCompleted(ctx, id)
	if err != nil {
		return nil, err
	}

	c.invalidatePayment(ctx, payment)
	return payment, nil
}

// MarkAsFailed marks a payment as failed and invalidates caches.
func (c *CachedPaymentRepository) MarkAsFailed(ctx context.Context, id int64, reason string) (*models.Payment, error) {
	payment, err := c.repo.MarkAsFailed(ctx, id, reason)
	if err != nil {
		return nil, err
	}

	c.invalidatePayment(ctx, payment)
	return payment, nil
}

// invalidatePayment invalidates all caches related to a payment.
func (c *CachedPaymentRepository) invalidatePayment(ctx context.Context, payment *models.Payment) {
	c.del(ctx, keyPaymentByID(payment.ID))
	c.del(ctx, keyPaymentByRef(payment.ReferenceID))
	c.invalidateUserLists(ctx, payment.UserID)
	c.invalidateAccountLists(ctx, payment.AccountID)
	c.invalidateGlobalLists(ctx)
}

// setCache marshals the value and stores it in Redis. Errors are logged, never returned.
func (c *CachedPaymentRepository) setCache(ctx context.Context, key string, value interface{}, ttl time.Duration) {
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
func (c *CachedPaymentRepository) del(ctx context.Context, keys ...string) {
	if err := c.redis.Del(ctx, keys...).Err(); err != nil {
		log.Printf("cache: failed to delete keys: %v", err)
	}
}

// invalidateUserLists removes cached list entries for a user using a pattern scan.
func (c *CachedPaymentRepository) invalidateUserLists(ctx context.Context, userID int64) {
	pattern := fmt.Sprintf("payment:user:%d:*", userID)
	c.deleteByPattern(ctx, pattern)
}

// invalidateAccountLists removes cached list entries for an account using a pattern scan.
func (c *CachedPaymentRepository) invalidateAccountLists(ctx context.Context, accountID int64) {
	pattern := fmt.Sprintf("payment:account:%d:*", accountID)
	c.deleteByPattern(ctx, pattern)
}

// invalidateGlobalLists removes all paginated list caches.
func (c *CachedPaymentRepository) invalidateGlobalLists(ctx context.Context) {
	c.deleteByPattern(ctx, "payment:all:*")
}

// deleteByPattern scans and deletes keys matching a pattern.
func (c *CachedPaymentRepository) deleteByPattern(ctx context.Context, pattern string) {
	iter := c.redis.Scan(ctx, 0, pattern, 100).Iterator()
	for iter.Next(ctx) {
		c.del(ctx, iter.Val())
	}
	if err := iter.Err(); err != nil {
		log.Printf("cache: failed to scan pattern %s: %v", pattern, err)
	}
}
