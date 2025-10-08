// path: backend/internal/social/adapter.go
package social

import (
	"context"
	"time"
)

// SocialAdapter defines the interface that all social platform integrations must implement
type SocialAdapter interface {
	// AuthRedirect generates the OAuth authorization URL for the platform
	AuthRedirect(ctx context.Context, state string, redirectURI string) (string, error)

	// HandleOAuthCallback processes the OAuth callback and returns access/refresh tokens
	HandleOAuthCallback(ctx context.Context, code string, redirectURI string) (*OAuthTokenResponse, error)

	// PostContent publishes content to the social platform
	PostContent(ctx context.Context, token *PlatformToken, content *PostContent) (*PostResult, error)

	// RefreshTokenIfNeeded checks token expiry and refreshes if needed
	RefreshTokenIfNeeded(ctx context.Context, token *PlatformToken) (*PlatformToken, error)

	// GetAccountInfo retrieves account information from the platform
	GetAccountInfo(ctx context.Context, token *PlatformToken) (*AccountInfo, error)

	// ValidateToken verifies if the token is still valid
	ValidateToken(ctx context.Context, token *PlatformToken) (bool, error)

	// RevokeToken revokes the access token
	RevokeToken(ctx context.Context, token *PlatformToken) error

	// GetRateLimits returns current rate limit status
	GetRateLimits(ctx context.Context, token *PlatformToken) (*RateLimitInfo, error)

	// GetPlatformName returns the platform identifier
	GetPlatformName() PlatformType

	// GetCapabilities returns what this platform supports
	GetCapabilities() *PlatformCapabilities
}

// PlatformType represents supported social platforms
type PlatformType string

const (
	PlatformTwitter   PlatformType = "twitter"
	PlatformFacebook  PlatformType = "facebook"
	PlatformInstagram PlatformType = "instagram"
	PlatformLinkedIn  PlatformType = "linkedin"
	PlatformTikTok    PlatformType = "tiktok"
	PlatformYouTube   PlatformType = "youtube"
)

// OAuthTokenResponse represents the response from OAuth flow
type OAuthTokenResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	ExpiresIn    int64     `json:"expires_in"`
	ExpiresAt    time.Time `json:"expires_at"`
	TokenType    string    `json:"token_type"`
	Scope        string    `json:"scope"`

	// Platform-specific data
	PlatformUserID   string                 `json:"platform_user_id"`
	PlatformUsername string                 `json:"platform_username"`
	Extra            map[string]interface{} `json:"extra,omitempty"`
}

// PlatformToken represents stored token data
type PlatformToken struct {
	ID               int64        `json:"id"`
	UserID           int64        `json:"user_id"`
	PlatformType     PlatformType `json:"platform_type"`
	PlatformUserID   string       `json:"platform_user_id"`
	PlatformUsername string       `json:"platform_username"`

	// Encrypted token data
	AccessToken  string `json:"-"` // Never expose in JSON
	RefreshToken string `json:"-"` // Never expose in JSON

	ExpiresAt     time.Time `json:"expires_at"`
	Scope         string    `json:"scope"`
	IsValid       bool      `json:"is_valid"`
	LastValidated time.Time `json:"last_validated"`

	// Metadata
	Extra     map[string]interface{} `json:"extra,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}

// PostContent represents content to be posted
type PostContent struct {
	Text        string     `json:"text"`
	MediaURLs   []string   `json:"media_urls,omitempty"`
	MediaType   MediaType  `json:"media_type,omitempty"`
	Link        string     `json:"link,omitempty"`
	ScheduledAt *time.Time `json:"scheduled_at,omitempty"`

	// Platform-specific options
	Options map[string]interface{} `json:"options,omitempty"`
}

type MediaType string

const (
	MediaTypeImage MediaType = "image"
	MediaTypeVideo MediaType = "video"
	MediaTypeGIF   MediaType = "gif"
)

// PostResult represents the result of a post operation
type PostResult struct {
	PlatformPostID string                 `json:"platform_post_id"`
	URL            string                 `json:"url"`
	PublishedAt    time.Time              `json:"published_at"`
	Success        bool                   `json:"success"`
	Error          string                 `json:"error,omitempty"`
	Extra          map[string]interface{} `json:"extra,omitempty"`
}

// AccountInfo represents social account information
type AccountInfo struct {
	PlatformUserID  string `json:"platform_user_id"`
	Username        string `json:"username"`
	DisplayName     string `json:"display_name"`
	ProfileImageURL string `json:"profile_image_url"`
	FollowersCount  int64  `json:"followers_count"`
	FollowingCount  int64  `json:"following_count"`
	IsVerified      bool   `json:"is_verified"`
	Bio             string `json:"bio,omitempty"`
}

// RateLimitInfo represents rate limit status
type RateLimitInfo struct {
	Limit     int       `json:"limit"`
	Remaining int       `json:"remaining"`
	ResetAt   time.Time `json:"reset_at"`
}

// PlatformCapabilities defines what features a platform supports
type PlatformCapabilities struct {
	SupportsImages        bool `json:"supports_images"`
	SupportsVideos        bool `json:"supports_videos"`
	SupportsMultipleMedia bool `json:"supports_multiple_media"`
	SupportsScheduling    bool `json:"supports_scheduling"`
	SupportsHashtags      bool `json:"supports_hashtags"`
	MaxTextLength         int  `json:"max_text_length"`
	MaxMediaCount         int  `json:"max_media_count"`
	MaxVideoSizeMB        int  `json:"max_video_size_mb"`
}
