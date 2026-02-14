package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"account/models"
	"account/repository"

	"github.com/redis/go-redis/v9"
	"github.com/shopspring/decimal"
)

const (
	accountByIDTTL   = 3 * time.Minute
	accountByNumTTL  = 3 * time.Minute
	accountUserTTL   = 2 * time.Minute
	accountActiveTTL = 1 * time.Minute
	accountAllTTL    = 1 * time.Minute
)

// CachedAccountRepository wraps an AccountRepository with Redis caching.
type CachedAccountRepository struct {
	repo  *repository.AccountRepository
	redis *redis.Client
}

// NewCachedAccountRepository creates a new cached account repository.
func NewCachedAccountRepository(repo *repository.AccountRepository, redisClient *redis.Client) *CachedAccountRepository {
	return &CachedAccountRepository{
		repo:  repo,
		redis: redisClient,
	}
}

func keyAccountByID(id int64) string {
	return fmt.Sprintf("account:id:%d", id)
}

func keyAccountByNum(number string) string {
	return fmt.Sprintf("account:num:%s", number)
}

func keyAccountByUser(userID int64, limit, offset int) string {
	return fmt.Sprintf("account:user:%d:%d:%d", userID, limit, offset)
}

func keyAccountActive() string {
	return "account:active"
}

func keyAccountAll(limit, offset int) string {
	return fmt.Sprintf("account:all:%d:%d", limit, offset)
}

// Create delegates to the underlying repo and invalidates list caches for that user.
func (c *CachedAccountRepository) Create(ctx context.Context, req *models.CreateAccountRequest) (*models.Account, error) {
	account, err := c.repo.Create(ctx, req)
	if err != nil {
		return nil, err
	}
	c.invalidateUserLists(ctx, account.UserID)
	c.invalidateGlobalLists(ctx)
	return account, nil
}

// GetByID checks cache first, falls back to DB.
func (c *CachedAccountRepository) GetByID(ctx context.Context, id int64) (*models.Account, error) {
	key := keyAccountByID(id)

	data, err := c.redis.Get(ctx, key).Bytes()
	if err == nil {
		var account models.Account
		if json.Unmarshal(data, &account) == nil {
			return &account, nil
		}
	}

	account, err := c.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	c.setCache(ctx, key, account, accountByIDTTL)
	return account, nil
}

// GetByAccountNumber checks cache first, falls back to DB.
func (c *CachedAccountRepository) GetByAccountNumber(ctx context.Context, accountNumber string) (*models.Account, error) {
	key := keyAccountByNum(accountNumber)

	data, err := c.redis.Get(ctx, key).Bytes()
	if err == nil {
		var account models.Account
		if json.Unmarshal(data, &account) == nil {
			return &account, nil
		}
	}

	account, err := c.repo.GetByAccountNumber(ctx, accountNumber)
	if err != nil {
		return nil, err
	}

	c.setCache(ctx, key, account, accountByNumTTL)
	return account, nil
}

// ListByUserID checks cache first, falls back to DB.
func (c *CachedAccountRepository) ListByUserID(ctx context.Context, userID int64, limit, offset int) (*models.AccountListResponse, error) {
	key := keyAccountByUser(userID, limit, offset)

	data, err := c.redis.Get(ctx, key).Bytes()
	if err == nil {
		var result models.AccountListResponse
		if json.Unmarshal(data, &result) == nil {
			return &result, nil
		}
	}

	result, err := c.repo.ListByUserID(ctx, userID, limit, offset)
	if err != nil {
		return nil, err
	}

	c.setCache(ctx, key, result, accountUserTTL)
	return result, nil
}

// ListAllActive checks cache first, falls back to DB.
func (c *CachedAccountRepository) ListAllActive(ctx context.Context) (*models.AccountListResponse, error) {
	key := keyAccountActive()

	data, err := c.redis.Get(ctx, key).Bytes()
	if err == nil {
		var result models.AccountListResponse
		if json.Unmarshal(data, &result) == nil {
			return &result, nil
		}
	}

	result, err := c.repo.ListAllActive(ctx)
	if err != nil {
		return nil, err
	}

	c.setCache(ctx, key, result, accountActiveTTL)
	return result, nil
}

// ListAll checks cache first, falls back to DB.
func (c *CachedAccountRepository) ListAll(ctx context.Context, limit, offset int) (*models.AccountListResponse, error) {
	key := keyAccountAll(limit, offset)

	data, err := c.redis.Get(ctx, key).Bytes()
	if err == nil {
		var result models.AccountListResponse
		if json.Unmarshal(data, &result) == nil {
			return &result, nil
		}
	}

	result, err := c.repo.ListAll(ctx, limit, offset)
	if err != nil {
		return nil, err
	}

	c.setCache(ctx, key, result, accountAllTTL)
	return result, nil
}

// Update delegates to repo and invalidates affected caches.
func (c *CachedAccountRepository) Update(ctx context.Context, id int64, req *models.UpdateAccountRequest) (*models.Account, error) {
	// Get existing account to know the account number for cache invalidation
	existing, _ := c.repo.GetByID(ctx, id)

	account, err := c.repo.Update(ctx, id, req)
	if err != nil {
		return nil, err
	}

	c.invalidateAccount(ctx, id, account.AccountNumber)
	if existing != nil {
		c.invalidateUserLists(ctx, existing.UserID)
	}
	c.invalidateGlobalLists(ctx)
	return account, nil
}

// Delete delegates to repo and invalidates affected caches.
func (c *CachedAccountRepository) Delete(ctx context.Context, id int64) error {
	// Get existing account to know the account number for cache invalidation
	existing, _ := c.repo.GetByID(ctx, id)

	err := c.repo.Delete(ctx, id)
	if err != nil {
		return err
	}

	if existing != nil {
		c.invalidateAccount(ctx, id, existing.AccountNumber)
		c.invalidateUserLists(ctx, existing.UserID)
	}
	c.invalidateGlobalLists(ctx)
	return nil
}

// Deposit delegates to repo and invalidates affected caches.
func (c *CachedAccountRepository) Deposit(ctx context.Context, id int64, amount decimal.Decimal) (*models.Account, error) {
	account, err := c.repo.Deposit(ctx, id, amount)
	if err != nil {
		return nil, err
	}

	c.invalidateAccount(ctx, id, account.AccountNumber)
	c.del(ctx, keyAccountActive())
	return account, nil
}

// Withdraw delegates to repo and invalidates affected caches.
func (c *CachedAccountRepository) Withdraw(ctx context.Context, id int64, amount decimal.Decimal) (*models.Account, error) {
	account, err := c.repo.Withdraw(ctx, id, amount)
	if err != nil {
		return nil, err
	}

	c.invalidateAccount(ctx, id, account.AccountNumber)
	c.del(ctx, keyAccountActive())
	return account, nil
}

// Transfer delegates to repo and invalidates both accounts.
func (c *CachedAccountRepository) Transfer(ctx context.Context, fromID, toID int64, amount decimal.Decimal) error {
	// Get account numbers before transfer for cache invalidation
	fromAcct, _ := c.repo.GetByID(ctx, fromID)
	toAcct, _ := c.repo.GetByID(ctx, toID)

	err := c.repo.Transfer(ctx, fromID, toID, amount)
	if err != nil {
		return err
	}

	c.del(ctx, keyAccountByID(fromID))
	c.del(ctx, keyAccountByID(toID))
	if fromAcct != nil {
		c.del(ctx, keyAccountByNum(fromAcct.AccountNumber))
	}
	if toAcct != nil {
		c.del(ctx, keyAccountByNum(toAcct.AccountNumber))
	}
	c.del(ctx, keyAccountActive())
	return nil
}

// setCache marshals the value and stores it in Redis. Errors are logged, never returned.
func (c *CachedAccountRepository) setCache(ctx context.Context, key string, value interface{}, ttl time.Duration) {
	data, err := json.Marshal(value)
	if err != nil {
		log.Printf("cache: failed to marshal %s: %v", key, err)
		return
	}
	if err := c.redis.Set(ctx, key, data, ttl).Err(); err != nil {
		log.Printf("cache: failed to set %s: %v", key, err)
	}
}

// del deletes a key from Redis. Errors are logged, never returned.
func (c *CachedAccountRepository) del(ctx context.Context, keys ...string) {
	if err := c.redis.Del(ctx, keys...).Err(); err != nil {
		log.Printf("cache: failed to delete keys: %v", err)
	}
}

// invalidateAccount removes both ID and account number cache entries.
func (c *CachedAccountRepository) invalidateAccount(ctx context.Context, id int64, accountNumber string) {
	c.del(ctx, keyAccountByID(id))
	if accountNumber != "" {
		c.del(ctx, keyAccountByNum(accountNumber))
	}
}

// invalidateUserLists removes cached list entries for a user using a pattern scan.
func (c *CachedAccountRepository) invalidateUserLists(ctx context.Context, userID int64) {
	pattern := fmt.Sprintf("account:user:%d:*", userID)
	c.deleteByPattern(ctx, pattern)
}

// invalidateGlobalLists removes the active directory cache and all paginated list caches.
func (c *CachedAccountRepository) invalidateGlobalLists(ctx context.Context) {
	c.del(ctx, keyAccountActive())
	c.deleteByPattern(ctx, "account:all:*")
}

// deleteByPattern scans and deletes keys matching a pattern.
func (c *CachedAccountRepository) deleteByPattern(ctx context.Context, pattern string) {
	iter := c.redis.Scan(ctx, 0, pattern, 100).Iterator()
	for iter.Next(ctx) {
		c.del(ctx, iter.Val())
	}
	if err := iter.Err(); err != nil {
		log.Printf("cache: failed to scan pattern %s: %v", pattern, err)
	}
}
