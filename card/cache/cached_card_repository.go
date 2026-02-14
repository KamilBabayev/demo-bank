package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"card/models"
	"card/repository"

	"github.com/redis/go-redis/v9"
)

const (
	cardByIDTTL      = 3 * time.Minute
	cardAccountTTL   = 2 * time.Minute
	cardAllTTL       = 1 * time.Minute
)

// CachedCardRepository wraps a CardRepository with Redis caching.
type CachedCardRepository struct {
	repo  *repository.CardRepository
	redis *redis.Client
}

// NewCachedCardRepository creates a new cached card repository.
func NewCachedCardRepository(repo *repository.CardRepository, redisClient *redis.Client) *CachedCardRepository {
	return &CachedCardRepository{
		repo:  repo,
		redis: redisClient,
	}
}

func keyCardByID(id int64) string {
	return fmt.Sprintf("card:id:%d", id)
}

func keyCardByAccount(accountID int64, limit, offset int) string {
	return fmt.Sprintf("card:account:%d:%d:%d", accountID, limit, offset)
}

func keyCardAll(limit, offset int) string {
	return fmt.Sprintf("card:all:%d:%d", limit, offset)
}

// Create delegates to the underlying repo and invalidates list caches.
func (c *CachedCardRepository) Create(ctx context.Context, req *models.CreateCardRequest) (*models.Card, error) {
	card, err := c.repo.Create(ctx, req)
	if err != nil {
		return nil, err
	}
	c.invalidateAccountLists(ctx, card.AccountID)
	c.invalidateGlobalLists(ctx)
	return card, nil
}

// GetByID checks cache first, falls back to DB.
func (c *CachedCardRepository) GetByID(ctx context.Context, id int64) (*models.Card, error) {
	key := keyCardByID(id)

	data, err := c.redis.Get(ctx, key).Bytes()
	if err == nil {
		var card models.Card
		if json.Unmarshal(data, &card) == nil {
			return &card, nil
		}
	}

	card, err := c.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	c.setCache(ctx, key, card, cardByIDTTL)
	return card, nil
}

// ListByAccountID checks cache first, falls back to DB.
func (c *CachedCardRepository) ListByAccountID(ctx context.Context, accountID int64, limit, offset int) (*models.CardListResponse, error) {
	key := keyCardByAccount(accountID, limit, offset)

	data, err := c.redis.Get(ctx, key).Bytes()
	if err == nil {
		var result models.CardListResponse
		if json.Unmarshal(data, &result) == nil {
			return &result, nil
		}
	}

	result, err := c.repo.ListByAccountID(ctx, accountID, limit, offset)
	if err != nil {
		return nil, err
	}

	c.setCache(ctx, key, result, cardAccountTTL)
	return result, nil
}

// ListAll checks cache first, falls back to DB.
func (c *CachedCardRepository) ListAll(ctx context.Context, limit, offset int) (*models.CardListResponse, error) {
	key := keyCardAll(limit, offset)

	data, err := c.redis.Get(ctx, key).Bytes()
	if err == nil {
		var result models.CardListResponse
		if json.Unmarshal(data, &result) == nil {
			return &result, nil
		}
	}

	result, err := c.repo.ListAll(ctx, limit, offset)
	if err != nil {
		return nil, err
	}

	c.setCache(ctx, key, result, cardAllTTL)
	return result, nil
}

// ListByUserAccounts checks cache first, falls back to DB.
func (c *CachedCardRepository) ListByUserAccounts(ctx context.Context, accountIDs []int64, limit, offset int) (*models.CardListResponse, error) {
	// No caching for multi-account queries (complex key space); delegate directly.
	return c.repo.ListByUserAccounts(ctx, accountIDs, limit, offset)
}

// Update delegates to repo and invalidates affected caches.
func (c *CachedCardRepository) Update(ctx context.Context, id int64, req *models.UpdateCardRequest) (*models.Card, error) {
	existing, _ := c.repo.GetByID(ctx, id)

	card, err := c.repo.Update(ctx, id, req)
	if err != nil {
		return nil, err
	}

	c.del(ctx, keyCardByID(id))
	if existing != nil {
		c.invalidateAccountLists(ctx, existing.AccountID)
	}
	c.invalidateGlobalLists(ctx)
	return card, nil
}

// Block blocks a card and invalidates caches.
func (c *CachedCardRepository) Block(ctx context.Context, id int64) (*models.Card, error) {
	existing, _ := c.repo.GetByID(ctx, id)

	card, err := c.repo.Block(ctx, id)
	if err != nil {
		return nil, err
	}

	c.del(ctx, keyCardByID(id))
	if existing != nil {
		c.invalidateAccountLists(ctx, existing.AccountID)
	}
	c.invalidateGlobalLists(ctx)
	return card, nil
}

// Unblock unblocks a card and invalidates caches.
func (c *CachedCardRepository) Unblock(ctx context.Context, id int64) (*models.Card, error) {
	existing, _ := c.repo.GetByID(ctx, id)

	card, err := c.repo.Unblock(ctx, id)
	if err != nil {
		return nil, err
	}

	c.del(ctx, keyCardByID(id))
	if existing != nil {
		c.invalidateAccountLists(ctx, existing.AccountID)
	}
	c.invalidateGlobalLists(ctx)
	return card, nil
}

// Cancel cancels a card and invalidates caches.
func (c *CachedCardRepository) Cancel(ctx context.Context, id int64) error {
	existing, _ := c.repo.GetByID(ctx, id)

	err := c.repo.Cancel(ctx, id)
	if err != nil {
		return err
	}

	c.del(ctx, keyCardByID(id))
	if existing != nil {
		c.invalidateAccountLists(ctx, existing.AccountID)
	}
	c.invalidateGlobalLists(ctx)
	return nil
}

// SetPIN sets a card PIN and invalidates the card cache.
func (c *CachedCardRepository) SetPIN(ctx context.Context, id int64, pin string) error {
	err := c.repo.SetPIN(ctx, id, pin)
	if err != nil {
		return err
	}

	c.del(ctx, keyCardByID(id))
	return nil
}

// setCache marshals the value and stores it in Redis. Errors are logged, never returned.
func (c *CachedCardRepository) setCache(ctx context.Context, key string, value interface{}, ttl time.Duration) {
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
func (c *CachedCardRepository) del(ctx context.Context, keys ...string) {
	if err := c.redis.Del(ctx, keys...).Err(); err != nil {
		log.Printf("cache: failed to delete keys: %v", err)
	}
}

// invalidateAccountLists removes cached list entries for an account using a pattern scan.
func (c *CachedCardRepository) invalidateAccountLists(ctx context.Context, accountID int64) {
	pattern := fmt.Sprintf("card:account:%d:*", accountID)
	c.deleteByPattern(ctx, pattern)
}

// invalidateGlobalLists removes all paginated list caches.
func (c *CachedCardRepository) invalidateGlobalLists(ctx context.Context) {
	c.deleteByPattern(ctx, "card:all:*")
}

// deleteByPattern scans and deletes keys matching a pattern.
func (c *CachedCardRepository) deleteByPattern(ctx context.Context, pattern string) {
	iter := c.redis.Scan(ctx, 0, pattern, 100).Iterator()
	for iter.Next(ctx) {
		c.del(ctx, iter.Val())
	}
	if err := iter.Err(); err != nil {
		log.Printf("cache: failed to scan pattern %s: %v", pattern, err)
	}
}
