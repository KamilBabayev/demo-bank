package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"notification-service/models"
	"notification-service/repository"

	"github.com/redis/go-redis/v9"
)

const (
	notificationByIDTTL = 3 * time.Minute
	notificationUserTTL = 1 * time.Minute
	notificationAllTTL  = 1 * time.Minute
)

// CachedNotificationRepository wraps a NotificationRepository with Redis caching.
type CachedNotificationRepository struct {
	repo  *repository.NotificationRepository
	redis *redis.Client
}

// NewCachedNotificationRepository creates a new cached notification repository.
func NewCachedNotificationRepository(repo *repository.NotificationRepository, redisClient *redis.Client) *CachedNotificationRepository {
	return &CachedNotificationRepository{
		repo:  repo,
		redis: redisClient,
	}
}

func keyNotificationByID(id int64) string {
	return fmt.Sprintf("notification:id:%d", id)
}

func keyNotificationByUser(userID int64, limit, offset int) string {
	return fmt.Sprintf("notification:user:%d:%d:%d", userID, limit, offset)
}

func keyNotificationAll(limit, offset int) string {
	return fmt.Sprintf("notification:all:%d:%d", limit, offset)
}

// Create delegates to the underlying repo and invalidates list caches.
func (c *CachedNotificationRepository) Create(ctx context.Context, req *models.CreateNotificationRequest) (*models.Notification, error) {
	notification, err := c.repo.Create(ctx, req)
	if err != nil {
		return nil, err
	}
	c.invalidateUserLists(ctx, notification.UserID)
	c.invalidateGlobalLists(ctx)
	return notification, nil
}

// CreateFromEvent delegates to the underlying repo and invalidates list caches.
func (c *CachedNotificationRepository) CreateFromEvent(ctx context.Context, userID int64, notifType, channel, title, content string, metadata map[string]interface{}) (*models.Notification, error) {
	notification, err := c.repo.CreateFromEvent(ctx, userID, notifType, channel, title, content, metadata)
	if err != nil {
		return nil, err
	}
	c.invalidateUserLists(ctx, userID)
	c.invalidateGlobalLists(ctx)
	return notification, nil
}

// GetByID checks cache first, falls back to DB.
func (c *CachedNotificationRepository) GetByID(ctx context.Context, id int64) (*models.Notification, error) {
	key := keyNotificationByID(id)

	data, err := c.redis.Get(ctx, key).Bytes()
	if err == nil {
		var notification models.Notification
		if json.Unmarshal(data, &notification) == nil {
			return &notification, nil
		}
	}

	notification, err := c.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	c.setCache(ctx, key, notification, notificationByIDTTL)
	return notification, nil
}

// ListByUserID checks cache first, falls back to DB.
func (c *CachedNotificationRepository) ListByUserID(ctx context.Context, userID int64, limit, offset int) (*models.NotificationListResponse, error) {
	key := keyNotificationByUser(userID, limit, offset)

	data, err := c.redis.Get(ctx, key).Bytes()
	if err == nil {
		var result models.NotificationListResponse
		if json.Unmarshal(data, &result) == nil {
			return &result, nil
		}
	}

	result, err := c.repo.ListByUserID(ctx, userID, limit, offset)
	if err != nil {
		return nil, err
	}

	c.setCache(ctx, key, result, notificationUserTTL)
	return result, nil
}

// ListAll checks cache first, falls back to DB.
func (c *CachedNotificationRepository) ListAll(ctx context.Context, limit, offset int) (*models.NotificationListResponse, error) {
	key := keyNotificationAll(limit, offset)

	data, err := c.redis.Get(ctx, key).Bytes()
	if err == nil {
		var result models.NotificationListResponse
		if json.Unmarshal(data, &result) == nil {
			return &result, nil
		}
	}

	result, err := c.repo.ListAll(ctx, limit, offset)
	if err != nil {
		return nil, err
	}

	c.setCache(ctx, key, result, notificationAllTTL)
	return result, nil
}

// MarkAsRead delegates to repo and invalidates affected caches.
func (c *CachedNotificationRepository) MarkAsRead(ctx context.Context, id int64) (*models.Notification, error) {
	notification, err := c.repo.MarkAsRead(ctx, id)
	if err != nil {
		return nil, err
	}

	c.del(ctx, keyNotificationByID(id))
	c.invalidateUserLists(ctx, notification.UserID)
	c.invalidateGlobalLists(ctx)
	return notification, nil
}

// MarkAsSent delegates to repo and invalidates affected caches.
func (c *CachedNotificationRepository) MarkAsSent(ctx context.Context, id int64) (*models.Notification, error) {
	notification, err := c.repo.MarkAsSent(ctx, id)
	if err != nil {
		return nil, err
	}

	c.del(ctx, keyNotificationByID(id))
	c.invalidateUserLists(ctx, notification.UserID)
	c.invalidateGlobalLists(ctx)
	return notification, nil
}

// MarkAllAsReadForUser delegates to repo and invalidates affected caches.
func (c *CachedNotificationRepository) MarkAllAsReadForUser(ctx context.Context, userID int64) error {
	err := c.repo.MarkAllAsReadForUser(ctx, userID)
	if err != nil {
		return err
	}

	// Invalidate all notification caches for this user
	c.deleteByPattern(ctx, fmt.Sprintf("notification:id:*"))
	c.invalidateUserLists(ctx, userID)
	c.invalidateGlobalLists(ctx)
	return nil
}

// Delete delegates to repo and invalidates affected caches.
func (c *CachedNotificationRepository) Delete(ctx context.Context, id int64) error {
	// Get notification before deleting to know userID for cache invalidation
	existing, _ := c.repo.GetByID(ctx, id)

	err := c.repo.Delete(ctx, id)
	if err != nil {
		return err
	}

	c.del(ctx, keyNotificationByID(id))
	if existing != nil {
		c.invalidateUserLists(ctx, existing.UserID)
	}
	c.invalidateGlobalLists(ctx)
	return nil
}

// setCache marshals the value and stores it in Redis. Errors are logged, never returned.
func (c *CachedNotificationRepository) setCache(ctx context.Context, key string, value interface{}, ttl time.Duration) {
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
func (c *CachedNotificationRepository) del(ctx context.Context, keys ...string) {
	if err := c.redis.Del(ctx, keys...).Err(); err != nil {
		log.Printf("cache: failed to delete keys: %v", err)
	}
}

// invalidateUserLists removes cached list entries for a user using a pattern scan.
func (c *CachedNotificationRepository) invalidateUserLists(ctx context.Context, userID int64) {
	pattern := fmt.Sprintf("notification:user:%d:*", userID)
	c.deleteByPattern(ctx, pattern)
}

// invalidateGlobalLists removes all paginated list caches.
func (c *CachedNotificationRepository) invalidateGlobalLists(ctx context.Context) {
	c.deleteByPattern(ctx, "notification:all:*")
}

// deleteByPattern scans and deletes keys matching a pattern.
func (c *CachedNotificationRepository) deleteByPattern(ctx context.Context, pattern string) {
	iter := c.redis.Scan(ctx, 0, pattern, 100).Iterator()
	for iter.Next(ctx) {
		c.del(ctx, iter.Val())
	}
	if err := iter.Err(); err != nil {
		log.Printf("cache: failed to scan pattern %s: %v", pattern, err)
	}
}
