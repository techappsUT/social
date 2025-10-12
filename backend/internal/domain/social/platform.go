// path: backend/internal/domain/social/platform.go
// 🆕 NEW - Clean Architecture

package social

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// PlatformAdapter defines the interface for social media platform adapters
// This is implemented by infrastructure layer for each platform
type PlatformAdapter interface {
	// Authentication
	GetAuthorizationURL(state string) (string, error)
	ExchangeToken(ctx context.Context, code string) (*Credentials, error)
	RefreshToken(ctx context.Context, refreshToken string) (*Credentials, error)
	RevokeAccess(ctx context.Context, account *Account) error

	// Account operations
	GetProfile(ctx context.Context, account *Account) (*ProfileInfo, error)
	VerifyCredentials(ctx context.Context, account *Account) (bool, error)

	// Publishing
	PublishPost(ctx context.Context, account *Account, post *PostRequest) (*PostResult, error)
	DeletePost(ctx context.Context, account *Account, postID string) error
	EditPost(ctx context.Context, account *Account, postID string, content *PostRequest) error

	// Media operations
	UploadMedia(ctx context.Context, account *Account, media *MediaUpload) (*MediaResult, error)

	// Analytics
	GetPostAnalytics(ctx context.Context, account *Account, postID string) (*PostAnalytics, error)
	GetAccountAnalytics(ctx context.Context, account *Account, period time.Duration) (*AccountAnalytics, error)

	// Platform-specific features
	GetRateLimits(ctx context.Context, account *Account) (*RateLimits, error)
	GetPlatformFeatures(ctx context.Context, account *Account) ([]string, error)

	// Platform identification
	Name() string
	Platform() Platform
}

// PostRequest represents a request to publish a post
type PostRequest struct {
	Text        string
	MediaIDs    []string
	MediaURLs   []string
	Hashtags    []string
	Mentions    []string
	Link        string
	Location    *Location
	ScheduledAt *time.Time
	ReplyToID   string // For threads/replies
	Metadata    map[string]interface{}
}

// PostResult represents the result of publishing a post
type PostResult struct {
	PlatformPostID string
	URL            string
	PublishedAt    time.Time
	Success        bool
	Error          error
	RateLimitInfo  *RateLimitInfo
}

// MediaUpload represents media to be uploaded
type MediaUpload struct {
	Data     []byte
	MimeType string
	Filename string
	AltText  string
	Tags     []string
}

// MediaResult represents the result of media upload
type MediaResult struct {
	MediaID      string
	MediaURL     string
	ThumbnailURL string
	Type         string
	Size         int64
	Duration     int // For video in seconds
}

// Location represents geographic location for posts
type Location struct {
	Latitude  float64
	Longitude float64
	Name      string
	PlaceID   string // Platform-specific place ID
}

// PostAnalytics represents analytics for a single post
type PostAnalytics struct {
	PostID      string
	Impressions int
	Reach       int
	Clicks      int
	Likes       int
	Comments    int
	Shares      int
	Saves       int
	VideoViews  int
	Engagement  float64
	UpdatedAt   time.Time
}

// AccountAnalytics represents account-level analytics
type AccountAnalytics struct {
	AccountID            string
	Period               time.Duration
	FollowersGained      int
	FollowersLost        int
	TotalImpressions     int
	TotalReach           int
	TotalEngagement      int
	EngagementRate       float64
	TopPosts             []PostAnalytics
	AudienceDemographics map[string]interface{}
	UpdatedAt            time.Time
}

// RateLimitInfo provides rate limit status after an API call
type RateLimitInfo struct {
	Limit     int
	Remaining int
	ResetsAt  time.Time
	Category  string // e.g., "posts", "media", "reads"
}

// PlatformRegistry manages platform adapters
type PlatformRegistry interface {
	Register(platform Platform, adapter PlatformAdapter)
	Get(platform Platform) (PlatformAdapter, error)
	List() []Platform
	IsSupported(platform Platform) bool
}

// PlatformConfig holds platform-specific configuration
type PlatformConfig struct {
	Platform     Platform
	ClientID     string
	ClientSecret string
	RedirectURI  string
	Scopes       []string
	APIVersion   string
	BaseURL      string
	Timeout      time.Duration
	MaxRetries   int
	RateLimit    int // requests per second
}

// PlatformCapabilities defines what a platform supports
type PlatformCapabilities struct {
	SupportsVideo       bool
	SupportsImages      bool
	SupportsMultiMedia  bool
	SupportsThreads     bool
	SupportsScheduling  bool
	SupportsEditing     bool
	SupportsAnalytics   bool
	SupportsStories     bool
	SupportsPolls       bool
	SupportsLiveVideo   bool
	MaxTextLength       int
	MaxMediaFiles       int
	MaxVideoLength      int   // seconds
	MaxImageSize        int64 // bytes
	MaxVideoSize        int64 // bytes
	SupportedImageTypes []string
	SupportedVideoTypes []string
}

// GetPlatformCapabilities returns capabilities for a platform
func GetPlatformCapabilities(platform Platform) PlatformCapabilities {
	switch platform {
	case PlatformTwitter:
		return PlatformCapabilities{
			SupportsVideo:       true,
			SupportsImages:      true,
			SupportsMultiMedia:  true,
			SupportsThreads:     true,
			SupportsScheduling:  false,
			SupportsEditing:     false,
			SupportsAnalytics:   true,
			SupportsStories:     false,
			SupportsPolls:       true,
			SupportsLiveVideo:   false,
			MaxTextLength:       280,
			MaxMediaFiles:       4,
			MaxVideoLength:      140,
			MaxImageSize:        5 * 1024 * 1024,   // 5MB
			MaxVideoSize:        512 * 1024 * 1024, // 512MB
			SupportedImageTypes: []string{"image/jpeg", "image/png", "image/gif", "image/webp"},
			SupportedVideoTypes: []string{"video/mp4"},
		}
	case PlatformFacebook:
		return PlatformCapabilities{
			SupportsVideo:       true,
			SupportsImages:      true,
			SupportsMultiMedia:  true,
			SupportsThreads:     false,
			SupportsScheduling:  true,
			SupportsEditing:     true,
			SupportsAnalytics:   true,
			SupportsStories:     true,
			SupportsPolls:       true,
			SupportsLiveVideo:   true,
			MaxTextLength:       63206,
			MaxMediaFiles:       10,
			MaxVideoLength:      240 * 60,                // 240 minutes
			MaxImageSize:        10 * 1024 * 1024,        // 10MB
			MaxVideoSize:        10 * 1024 * 1024 * 1024, // 10GB
			SupportedImageTypes: []string{"image/jpeg", "image/png", "image/gif", "image/bmp", "image/tiff"},
			SupportedVideoTypes: []string{"video/mp4", "video/mov", "video/avi"},
		}
	case PlatformLinkedIn:
		return PlatformCapabilities{
			SupportsVideo:       true,
			SupportsImages:      true,
			SupportsMultiMedia:  true,
			SupportsThreads:     false,
			SupportsScheduling:  false,
			SupportsEditing:     false,
			SupportsAnalytics:   true,
			SupportsStories:     false,
			SupportsPolls:       true,
			SupportsLiveVideo:   true,
			MaxTextLength:       3000,
			MaxMediaFiles:       9,
			MaxVideoLength:      10 * 60,                // 10 minutes
			MaxImageSize:        10 * 1024 * 1024,       // 10MB
			MaxVideoSize:        5 * 1024 * 1024 * 1024, // 5GB
			SupportedImageTypes: []string{"image/jpeg", "image/png", "image/gif"},
			SupportedVideoTypes: []string{"video/mp4", "video/quicktime"},
		}
	case PlatformInstagram:
		return PlatformCapabilities{
			SupportsVideo:       true,
			SupportsImages:      true,
			SupportsMultiMedia:  true,
			SupportsThreads:     false,
			SupportsScheduling:  false,
			SupportsEditing:     false,
			SupportsAnalytics:   true,
			SupportsStories:     true,
			SupportsPolls:       false,
			SupportsLiveVideo:   true,
			MaxTextLength:       2200,
			MaxMediaFiles:       10,
			MaxVideoLength:      60,
			MaxImageSize:        8 * 1024 * 1024,   // 8MB
			MaxVideoSize:        100 * 1024 * 1024, // 100MB
			SupportedImageTypes: []string{"image/jpeg", "image/png"},
			SupportedVideoTypes: []string{"video/mp4", "video/mov"},
		}
	default:
		// Default capabilities
		return PlatformCapabilities{
			SupportsVideo:       false,
			SupportsImages:      true,
			SupportsMultiMedia:  false,
			SupportsThreads:     false,
			SupportsScheduling:  false,
			SupportsEditing:     false,
			SupportsAnalytics:   false,
			SupportsStories:     false,
			SupportsPolls:       false,
			SupportsLiveVideo:   false,
			MaxTextLength:       5000,
			MaxMediaFiles:       1,
			MaxVideoLength:      0,
			MaxImageSize:        5 * 1024 * 1024,
			MaxVideoSize:        0,
			SupportedImageTypes: []string{"image/jpeg", "image/png"},
			SupportedVideoTypes: []string{},
		}
	}
}

// PlatformError represents a platform-specific error
type PlatformError struct {
	Platform   Platform
	Code       string
	Message    string
	Retry      bool
	RetryAfter *time.Time
}

func (e PlatformError) Error() string {
	return string(e.Platform) + ": " + e.Message
}

// WebhookEvent represents an incoming webhook from a platform
type WebhookEvent struct {
	ID          uuid.UUID
	Platform    Platform
	AccountID   uuid.UUID
	EventType   string
	Payload     map[string]interface{}
	ReceivedAt  time.Time
	ProcessedAt *time.Time
	Success     bool
	Error       string
}

// WebhookHandler processes platform webhooks
type WebhookHandler interface {
	HandleWebhook(ctx context.Context, platform Platform, payload []byte) error
	VerifySignature(platform Platform, payload []byte, signature string) bool
}
