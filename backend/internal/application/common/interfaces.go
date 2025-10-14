// path: backend/internal/application/common/interfaces.go
package common

import (
	"context"
	"time"
)

// ============================================================================
// CORE SERVICES
// ============================================================================

// TokenService handles JWT token operations
type TokenService interface {
	GenerateAccessToken(userID, email, role string) (string, error)
	GenerateRefreshToken(userID string) (string, error)
	ValidateAccessToken(token string) (*TokenClaims, error)
	ValidateRefreshToken(token string) (*TokenClaims, error) // ✅ ADD THIS
	RevokeRefreshToken(ctx context.Context, token string) error
}

// TokenClaims represents JWT token claims
type TokenClaims struct {
	UserID    string    `json:"userId"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	TeamID    string    `json:"teamId,omitempty"`
	ExpiresAt time.Time `json:"exp"`
	IssuedAt  time.Time `json:"iat"`
}

// EmailService handles email operations
type EmailService interface {
	SendVerificationEmail(ctx context.Context, email, token string) error
	SendPasswordResetEmail(ctx context.Context, email, token string) error
	SendWelcomeEmail(ctx context.Context, email, firstName string) error
	SendInvitationEmail(ctx context.Context, email, teamName, inviteToken string) error
}

// CacheService handles caching operations
type CacheService interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
}

// EventBus handles domain events
type EventBus interface {
	Publish(ctx context.Context, event Event) error
	Subscribe(eventType string, handler EventHandler) error
}

// Event represents a domain event
type Event interface {
	Type() string
	OccurredAt() time.Time
	AggregateID() string
}

// EventHandler processes events
type EventHandler func(ctx context.Context, event Event) error

// Logger handles structured logging
type Logger interface {
	Debug(msg string, fields ...interface{})
	Info(msg string, fields ...interface{})
	Warn(msg string, fields ...interface{})
	Error(msg string, fields ...interface{})
}
