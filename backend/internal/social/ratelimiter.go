// path: backend/internal/social/ratelimiter.go
package social

import (
	"context"
	"fmt"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// RateLimiter manages rate limiting for social platform API calls
type RateLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		limiters: make(map[string]*rate.Limiter),
	}
}

// GetLimiter returns a rate limiter for a specific platform+account
func (rl *RateLimiter) GetLimiter(platform PlatformType, accountID string) *rate.Limiter {
	key := fmt.Sprintf("%s:%s", platform, accountID)

	rl.mu.RLock()
	limiter, exists := rl.limiters[key]
	rl.mu.RUnlock()

	if exists {
		return limiter
	}

	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Double-check after acquiring write lock
	if limiter, exists := rl.limiters[key]; exists {
		return limiter
	}

	// Platform-specific rate limits
	var r rate.Limit
	var burst int

	switch platform {
	case PlatformTwitter:
		r = rate.Every(15 * time.Minute / 300) // 300 requests per 15 minutes
		burst = 10
	case PlatformFacebook:
		r = rate.Every(time.Hour / 200) // 200 requests per hour
		burst = 20
	case PlatformLinkedIn:
		r = rate.Every(24 * time.Hour / 100) // 100 requests per day
		burst = 5
	default:
		r = rate.Every(time.Minute / 60) // Default: 60 requests per minute
		burst = 10
	}

	limiter = rate.NewLimiter(r, burst)
	rl.limiters[key] = limiter

	return limiter
}

// Wait blocks until rate limit allows the request
func (rl *RateLimiter) Wait(ctx context.Context, platform PlatformType, accountID string) error {
	limiter := rl.GetLimiter(platform, accountID)
	return limiter.Wait(ctx)
}

// Allow checks if a request is allowed without blocking
func (rl *RateLimiter) Allow(platform PlatformType, accountID string) bool {
	limiter := rl.GetLimiter(platform, accountID)
	return limiter.Allow()
}
