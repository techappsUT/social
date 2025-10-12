// -------------------------------------------------------------------
// path: backend/internal/infrastructure/services/cache_service.go
package services

import (
	"context"
	"sync"
	"time"

	"github.com/techappsUT/social-queue/internal/application/common"
)

// InMemoryCacheService implements common.CacheService using in-memory storage
type InMemoryCacheService struct {
	data map[string]cacheItem
	mu   sync.RWMutex
}

type cacheItem struct {
	value     string
	expiresAt time.Time
}

// NewInMemoryCacheService creates a new in-memory cache service
func NewInMemoryCacheService() common.CacheService {
	cache := &InMemoryCacheService{
		data: make(map[string]cacheItem),
	}

	// Start cleanup goroutine
	go cache.cleanupExpired()

	return cache
}

// Get retrieves a value from cache
func (c *InMemoryCacheService) Get(ctx context.Context, key string) (string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, exists := c.data[key]
	if !exists {
		return "", nil
	}

	// Check if expired
	if time.Now().After(item.expiresAt) {
		return "", nil
	}

	return item.value, nil
}

// Set stores a value in cache with TTL
func (c *InMemoryCacheService) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data[key] = cacheItem{
		value:     value,
		expiresAt: time.Now().Add(ttl),
	}

	return nil
}

// Delete removes a value from cache
func (c *InMemoryCacheService) Delete(ctx context.Context, key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.data, key)
	return nil
}

// Exists checks if a key exists in cache
func (c *InMemoryCacheService) Exists(ctx context.Context, key string) (bool, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, exists := c.data[key]
	if !exists {
		return false, nil
	}

	// Check if expired
	if time.Now().After(item.expiresAt) {
		return false, nil
	}

	return true, nil
}

// cleanupExpired periodically removes expired items
func (c *InMemoryCacheService) cleanupExpired() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		now := time.Now()
		for key, item := range c.data {
			if now.After(item.expiresAt) {
				delete(c.data, key)
			}
		}
		c.mu.Unlock()
	}
}
