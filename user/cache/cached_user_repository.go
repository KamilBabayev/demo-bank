package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"user-service/models"
	"user-service/repository"

	"github.com/redis/go-redis/v9"
)

const (
	userByIDTTL    = 5 * time.Minute
	userByNameTTL  = 5 * time.Minute
	userByEmailTTL = 5 * time.Minute
	userListTTL    = 2 * time.Minute
)

// CachedUserRepository wraps a UserRepository with Redis caching.
type CachedUserRepository struct {
	repo  *repository.UserRepository
	redis *redis.Client
}

// NewCachedUserRepository creates a new cached user repository.
func NewCachedUserRepository(repo *repository.UserRepository, redisClient *redis.Client) *CachedUserRepository {
	return &CachedUserRepository{
		repo:  repo,
		redis: redisClient,
	}
}

func keyUserByID(id int64) string {
	return fmt.Sprintf("user:id:%d", id)
}

func keyUserByName(username string) string {
	return fmt.Sprintf("user:name:%s", username)
}

func keyUserByEmail(email string) string {
	return fmt.Sprintf("user:email:%s", email)
}

func keyUserList(limit, offset int) string {
	return fmt.Sprintf("user:list:%d:%d", limit, offset)
}

// Create delegates to the underlying repo and invalidates list caches.
func (c *CachedUserRepository) Create(ctx context.Context, req *models.CreateUserRequest, passwordHash string) (*models.User, error) {
	user, err := c.repo.Create(ctx, req, passwordHash)
	if err != nil {
		return nil, err
	}
	c.invalidateLists(ctx)
	return user, nil
}

// GetByID checks cache first, falls back to DB.
func (c *CachedUserRepository) GetByID(ctx context.Context, id int64) (*models.User, error) {
	key := keyUserByID(id)

	data, err := c.redis.Get(ctx, key).Bytes()
	if err == nil {
		var user models.User
		if json.Unmarshal(data, &user) == nil {
			return &user, nil
		}
	}

	user, err := c.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	c.setCache(ctx, key, user, userByIDTTL)
	return user, nil
}

// GetByUsername checks cache first, falls back to DB.
func (c *CachedUserRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	key := keyUserByName(username)

	data, err := c.redis.Get(ctx, key).Bytes()
	if err == nil {
		var user models.User
		if json.Unmarshal(data, &user) == nil {
			return &user, nil
		}
	}

	user, err := c.repo.GetByUsername(ctx, username)
	if err != nil {
		return nil, err
	}

	c.setCache(ctx, key, user, userByNameTTL)
	return user, nil
}

// GetByEmail checks cache first, falls back to DB.
func (c *CachedUserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	key := keyUserByEmail(email)

	data, err := c.redis.Get(ctx, key).Bytes()
	if err == nil {
		var user models.User
		if json.Unmarshal(data, &user) == nil {
			return &user, nil
		}
	}

	user, err := c.repo.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	c.setCache(ctx, key, user, userByEmailTTL)
	return user, nil
}

// List checks cache first, falls back to DB.
func (c *CachedUserRepository) List(ctx context.Context, limit, offset int) (*models.UserListResponse, error) {
	key := keyUserList(limit, offset)

	data, err := c.redis.Get(ctx, key).Bytes()
	if err == nil {
		var result models.UserListResponse
		if json.Unmarshal(data, &result) == nil {
			return &result, nil
		}
	}

	result, err := c.repo.List(ctx, limit, offset)
	if err != nil {
		return nil, err
	}

	c.setCache(ctx, key, result, userListTTL)
	return result, nil
}

// Update delegates to repo and invalidates affected caches.
func (c *CachedUserRepository) Update(ctx context.Context, id int64, req *models.UpdateUserRequest) (*models.User, error) {
	// Get existing user to know username/email for cache invalidation
	existing, _ := c.repo.GetByID(ctx, id)

	user, err := c.repo.Update(ctx, id, req)
	if err != nil {
		return nil, err
	}

	c.invalidateUser(ctx, id, existing)
	c.invalidateLists(ctx)
	return user, nil
}

// Delete delegates to repo and invalidates affected caches.
func (c *CachedUserRepository) Delete(ctx context.Context, id int64) error {
	// Get existing user to know username/email for cache invalidation
	existing, _ := c.repo.GetByID(ctx, id)

	err := c.repo.Delete(ctx, id)
	if err != nil {
		return err
	}

	c.invalidateUser(ctx, id, existing)
	c.invalidateLists(ctx)
	return nil
}

// UpdateLastLogin delegates to repo and invalidates affected caches.
func (c *CachedUserRepository) UpdateLastLogin(ctx context.Context, id int64) error {
	err := c.repo.UpdateLastLogin(ctx, id)
	if err != nil {
		return err
	}

	// Invalidate by ID; username/email caches will expire naturally
	c.del(ctx, keyUserByID(id))
	return nil
}

// IncrementFailedLoginAttempts passes through to the underlying repo (security fields are not cached).
func (c *CachedUserRepository) IncrementFailedLoginAttempts(ctx context.Context, username string, maxAttempts int, lockDuration time.Duration) error {
	return c.repo.IncrementFailedLoginAttempts(ctx, username, maxAttempts, lockDuration)
}

// setCache marshals the value and stores it in Redis. Errors are logged, never returned.
func (c *CachedUserRepository) setCache(ctx context.Context, key string, value interface{}, ttl time.Duration) {
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
func (c *CachedUserRepository) del(ctx context.Context, keys ...string) {
	if err := c.redis.Del(ctx, keys...).Err(); err != nil {
		log.Printf("cache: failed to delete keys: %v", err)
	}
}

// invalidateUser removes all cached entries for a user.
func (c *CachedUserRepository) invalidateUser(ctx context.Context, id int64, existing *models.User) {
	c.del(ctx, keyUserByID(id))
	if existing != nil {
		c.del(ctx, keyUserByName(existing.Username))
		c.del(ctx, keyUserByEmail(existing.Email))
	}
}

// invalidateLists removes all cached user list entries.
func (c *CachedUserRepository) invalidateLists(ctx context.Context) {
	iter := c.redis.Scan(ctx, 0, "user:list:*", 100).Iterator()
	for iter.Next(ctx) {
		c.del(ctx, iter.Val())
	}
	if err := iter.Err(); err != nil {
		log.Printf("cache: failed to scan user:list:* pattern: %v", err)
	}
}
