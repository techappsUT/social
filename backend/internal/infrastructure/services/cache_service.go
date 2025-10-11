package services

import (
	"context"
	"sync"
	"time"

	"github.com/techappsUT/social-queue/internal/application/common"
)

type InMemoryCacheService struct {
	data map[string]string
	mu   sync.RWMutex
}

func NewInMemoryCacheService() common.CacheService {
	return &InMemoryCacheService{
		data: make(map[string]string),
	}
}

func (c *InMemoryCacheService) Get(ctx context.Context, key string) (string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.data[key], nil
}

func (c *InMemoryCacheService) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[key] = value
	return nil
}

func (c *InMemoryCacheService) Delete(ctx context.Context, key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.data, key)
	return nil
}

func (c *InMemoryCacheService) Exists(ctx context.Context, key string) (bool, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	_, exists := c.data[key]
	return exists, nil
}
