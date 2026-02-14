package cache

import (
	"context"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

// NewRedisClient creates a Redis client from a URL string.
func NewRedisClient(redisURL string) *redis.Client {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Printf("WARNING: Failed to parse Redis URL: %v, using defaults", err)
		opts = &redis.Options{
			Addr: "localhost:6379",
		}
	}

	opts.PoolSize = 10
	opts.DialTimeout = 3 * time.Second
	opts.ReadTimeout = 2 * time.Second
	opts.WriteTimeout = 2 * time.Second

	client := redis.NewClient(opts)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		log.Printf("WARNING: Redis is unavailable: %v (will fall back to database)", err)
	} else {
		log.Println("Redis connection established")
	}

	return client
}

// HealthCheck checks if Redis is reachable.
func HealthCheck(ctx context.Context, client *redis.Client) error {
	return client.Ping(ctx).Err()
}
