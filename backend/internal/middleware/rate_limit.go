// ============================================================================
// FILE: backend/internal/middleware/rate_limit.go
// PURPOSE: Redis-based rate limiting middleware with sliding window algorithm
// ============================================================================

package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/techappsUT/social-queue/internal/application/common"
)

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	// Requests per window
	RequestsPerWindow int
	// Window duration
	WindowDuration time.Duration
	// Key prefix for Redis
	KeyPrefix string
}

// DefaultRateLimitConfigs provides sensible defaults
var DefaultRateLimitConfigs = map[string]RateLimitConfig{
	"user": {
		RequestsPerWindow: 100,
		WindowDuration:    time.Minute,
		KeyPrefix:         "ratelimit:user",
	},
	"ip": {
		RequestsPerWindow: 1000,
		WindowDuration:    time.Minute,
		KeyPrefix:         "ratelimit:ip",
	},
	"auth": {
		RequestsPerWindow: 10,
		WindowDuration:    time.Minute,
		KeyPrefix:         "ratelimit:auth",
	},
}

// RateLimiter implements rate limiting using Redis
type RateLimiter struct {
	redis  *redis.Client
	logger common.Logger
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(redis *redis.Client, logger common.Logger) *RateLimiter {
	return &RateLimiter{
		redis:  redis,
		logger: logger,
	}
}

// RateLimitByIP limits requests per IP address
func (rl *RateLimiter) RateLimitByIP(config RateLimitConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := extractIP(r)
			key := fmt.Sprintf("%s:%s", config.KeyPrefix, ip)

			allowed, remaining, resetAt, err := rl.checkRateLimit(r.Context(), key, config)
			if err != nil {
				rl.logger.Error(fmt.Sprintf("Rate limit check failed: %v", err))
				// On error, allow the request (fail open)
				next.ServeHTTP(w, r)
				return
			}

			// Set rate limit headers
			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(config.RequestsPerWindow))
			w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
			w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(resetAt.Unix(), 10))

			if !allowed {
				w.Header().Set("Retry-After", strconv.FormatInt(int64(time.Until(resetAt).Seconds()), 10))
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				response := map[string]interface{}{
					"error":      "Rate limit exceeded",
					"message":    fmt.Sprintf("Too many requests. Please try again in %d seconds.", int(time.Until(resetAt).Seconds())),
					"retryAfter": int(time.Until(resetAt).Seconds()),
				}
				json.NewEncoder(w).Encode(response)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RateLimitByUser limits requests per authenticated user
func (rl *RateLimiter) RateLimitByUser(config RateLimitConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract user ID from context (set by auth middleware)
			userID, ok := GetUserID(r.Context())
			if !ok {
				// If no user ID, skip rate limiting (not authenticated)
				next.ServeHTTP(w, r)
				return
			}

			key := fmt.Sprintf("%s:%s", config.KeyPrefix, userID.String())
			allowed, remaining, resetAt, err := rl.checkRateLimit(r.Context(), key, config)
			if err != nil {
				rl.logger.Error(fmt.Sprintf("Rate limit check failed: %v", err))
				// On error, allow the request (fail open)
				next.ServeHTTP(w, r)
				return
			}

			// Set rate limit headers
			w.Header().Set("X-RateLimit-Limit", strconv.Itoa(config.RequestsPerWindow))
			w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
			w.Header().Set("X-RateLimit-Reset", strconv.FormatInt(resetAt.Unix(), 10))

			if !allowed {
				w.Header().Set("Retry-After", strconv.FormatInt(int64(time.Until(resetAt).Seconds()), 10))
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				response := map[string]interface{}{
					"error":      "Rate limit exceeded",
					"message":    fmt.Sprintf("Too many requests. Please try again in %d seconds.", int(time.Until(resetAt).Seconds())),
					"retryAfter": int(time.Until(resetAt).Seconds()),
				}
				json.NewEncoder(w).Encode(response)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// checkRateLimit implements sliding window rate limiting using Redis
func (rl *RateLimiter) checkRateLimit(ctx context.Context, key string, config RateLimitConfig) (allowed bool, remaining int, resetAt time.Time, err error) {
	now := time.Now()
	windowStart := now.Add(-config.WindowDuration)

	// Use Redis sorted set for sliding window
	// Score = timestamp, Value = unique request ID

	pipe := rl.redis.Pipeline()

	// Remove old entries outside the window
	pipe.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", windowStart.UnixNano()))

	// Count current entries in window
	countCmd := pipe.ZCard(ctx, key)

	// Add current request
	pipe.ZAdd(ctx, key, redis.Z{
		Score:  float64(now.UnixNano()),
		Member: fmt.Sprintf("%d", now.UnixNano()),
	})

	// Set expiry on the key
	pipe.Expire(ctx, key, config.WindowDuration+time.Minute)

	// Execute pipeline
	_, err = pipe.Exec(ctx)
	if err != nil {
		return false, 0, time.Time{}, fmt.Errorf("redis pipeline failed: %w", err)
	}

	count := int(countCmd.Val())

	// Check if rate limit exceeded
	if count >= config.RequestsPerWindow {
		remaining = 0
		resetAt = now.Add(config.WindowDuration)
		return false, remaining, resetAt, nil
	}

	remaining = config.RequestsPerWindow - count - 1 // -1 for current request
	resetAt = now.Add(config.WindowDuration)
	return true, remaining, resetAt, nil
}

// extractIP extracts the real IP address from the request
// func extractIP(r *http.Request) string {
// 	// Check X-Forwarded-For header (behind proxy/load balancer)
// 	xff := r.Header.Get("X-Forwarded-For")
// 	if xff != "" {
// 		// Take the first IP in the list
// 		ips := strings.Split(xff, ",")
// 		if len(ips) > 0 {
// 			return strings.TrimSpace(ips[0])
// 		}
// 	}

// 	// Check X-Real-IP header
// 	xri := r.Header.Get("X-Real-IP")
// 	if xri != "" {
// 		return strings.TrimSpace(xri)
// 	}

// 	// Fall back to RemoteAddr
// 	ip := r.RemoteAddr
// 	// Remove port if present
// 	if idx := strings.LastIndex(ip, ":"); idx != -1 {
// 		ip = ip[:idx]
// 	}

// 	return ip
// }

// ClearRateLimit clears rate limit for a specific key (admin function)
func (rl *RateLimiter) ClearRateLimit(ctx context.Context, key string) error {
	return rl.redis.Del(ctx, key).Err()
}

// GetRateLimitStatus returns current rate limit status for a key
func (rl *RateLimiter) GetRateLimitStatus(ctx context.Context, key string, config RateLimitConfig) (count int, resetAt time.Time, err error) {
	now := time.Now()
	windowStart := now.Add(-config.WindowDuration)

	// Count entries in current window
	cnt, err := rl.redis.ZCount(ctx, key, fmt.Sprintf("%d", windowStart.UnixNano()), "+inf").Result()
	if err != nil {
		return 0, time.Time{}, fmt.Errorf("failed to get rate limit status: %w", err)
	}

	return int(cnt), now.Add(config.WindowDuration), nil
}
