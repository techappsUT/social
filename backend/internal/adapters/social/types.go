// ============================================================================
// FILE: backend/internal/adapters/social/types.go
// FIXED VERSION - Added PlatformUserID to Token
// ============================================================================
package social

import (
	"context"
	"time"
)

// Adapter defines the interface that all social platform adapters must implement
type Adapter interface {
	// OAuth methods
	GetAuthURL(state string, scopes []string) string
	ExchangeCode(ctx context.Context, code string) (*Token, error)
	RefreshToken(ctx context.Context, refreshToken string) (*Token, error)
	ValidateToken(ctx context.Context, token *Token) (bool, error)

	// Publishing methods
	PublishPost(ctx context.Context, token *Token, content *PostContent) (*PublishResult, error)

	// Analytics methods
	GetPostAnalytics(ctx context.Context, token *Token, postID string) (*Analytics, error)
}

// Token represents OAuth tokens
type Token struct {
	AccessToken    string
	RefreshToken   string
	ExpiresAt      *time.Time
	Scopes         []string
	PlatformUserID string // ADDED: Platform-specific user ID
}

// PostContent represents content to be published
type PostContent struct {
	Text      string
	MediaURLs []string
	Link      string
	Metadata  map[string]interface{}
}

// PublishResult represents the result of publishing a post
type PublishResult struct {
	PlatformPostID string
	URL            string
	PublishedAt    time.Time
}

// Analytics represents post analytics/metrics
type Analytics struct {
	Impressions int
	Engagements int
	Likes       int
	Shares      int
	Comments    int
	Clicks      int
}

// UserInfo represents basic user information from OAuth
type UserInfo struct {
	ID          string
	Username    string
	DisplayName string
	Email       string
	AvatarURL   string
}
