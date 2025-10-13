// ============================================================================
// FILE: backend/internal/infrastructure/services/redis_cache.go
// PURPOSE: Redis-based cache service with distributed locking
// ============================================================================

package services

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/techappsUT/social-queue/internal/application/common"
)

// RedisCacheService implements common.CacheService using Redis
type RedisCacheService struct {
	client *redis.Client
	logger common.Logger
}

// NewRedisCacheService creates a new Redis cache service
func NewRedisCacheService(host string, port int, password string, db int, logger common.Logger) (common.CacheService, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", host, port),
		Password:     password,
		DB:           db,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolSize:     10,
		MinIdleConns: 2,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	logger.Info("Redis cache service initialized successfully")

	return &RedisCacheService{
		client: client,
		logger: logger,
	}, nil
}

// Get retrieves a value from Redis
func (r *RedisCacheService) Get(ctx context.Context, key string) (string, error) {
	val, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil // Key doesn't exist
	}
	if err != nil {
		return "", fmt.Errorf("redis get failed: %w", err)
	}
	return val, nil
}

// Set stores a value in Redis with TTL
func (r *RedisCacheService) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	if err := r.client.Set(ctx, key, value, ttl).Err(); err != nil {
		return fmt.Errorf("redis set failed: %w", err)
	}
	return nil
}

// Delete removes a key from Redis
func (r *RedisCacheService) Delete(ctx context.Context, key string) error {
	if err := r.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("redis delete failed: %w", err)
	}
	return nil
}

// Exists checks if a key exists in Redis
func (r *RedisCacheService) Exists(ctx context.Context, key string) (bool, error) {
	count, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("redis exists failed: %w", err)
	}
	return count > 0, nil
}

// Lock acquires a distributed lock using Redis SET NX with TTL
func (r *RedisCacheService) Lock(ctx context.Context, key string, ttl time.Duration) (bool, error) {
	lockKey := fmt.Sprintf("lock:%s", key)
	acquired, err := r.client.SetNX(ctx, lockKey, "1", ttl).Result()
	if err != nil {
		return false, fmt.Errorf("redis lock failed: %w", err)
	}
	return acquired, nil
}

// Unlock releases a distributed lock
func (r *RedisCacheService) Unlock(ctx context.Context, key string) error {
	lockKey := fmt.Sprintf("lock:%s", key)
	if err := r.client.Del(ctx, lockKey).Err(); err != nil {
		return fmt.Errorf("redis unlock failed: %w", err)
	}
	return nil
}

// Increment atomically increments a value (for rate limiting)
func (r *RedisCacheService) Increment(ctx context.Context, key string, ttl time.Duration) (int64, error) {
	pipe := r.client.Pipeline()
	incr := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, ttl)

	if _, err := pipe.Exec(ctx); err != nil {
		return 0, fmt.Errorf("redis increment failed: %w", err)
	}

	return incr.Val(), nil
}

// Close closes the Redis connection
func (r *RedisCacheService) Close() error {
	if err := r.client.Close(); err != nil {
		return fmt.Errorf("failed to close Redis connection: %w", err)
	}
	r.logger.Info("Redis connection closed")
	return nil
}

// Client returns the underlying Redis client (for advanced usage like worker queues)
func (r *RedisCacheService) Client() *redis.Client {
	return r.client
}

// Health checks Redis connection health
func (r *RedisCacheService) Health(ctx context.Context) error {
	if err := r.client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("redis health check failed: %w", err)
	}
	return nil
}
