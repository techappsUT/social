// path: backend/internal/domain/social/repository.go
// 🆕 NEW - Clean Architecture

package social

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// AccountRepository defines the interface for social account persistence
type AccountRepository interface {
	// CRUD operations
	Create(ctx context.Context, account *Account) error
	Update(ctx context.Context, account *Account) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*Account, error)

	// Team-based queries
	FindByTeamID(ctx context.Context, teamID uuid.UUID) ([]*Account, error)
	FindByTeamAndPlatform(ctx context.Context, teamID uuid.UUID, platform Platform) ([]*Account, error)
	CountByTeamID(ctx context.Context, teamID uuid.UUID) (int64, error)

	// User-based queries
	FindByUserID(ctx context.Context, userID uuid.UUID) ([]*Account, error)

	// Platform queries
	FindByPlatform(ctx context.Context, platform Platform, offset, limit int) ([]*Account, error)
	FindByPlatformUserID(ctx context.Context, platform Platform, platformUserID string) (*Account, error)

	// Status queries
	FindByStatus(ctx context.Context, status Status, offset, limit int) ([]*Account, error)
	FindExpiredAccounts(ctx context.Context) ([]*Account, error)
	FindExpiringAccounts(ctx context.Context, withinDays int) ([]*Account, error)
	FindRateLimitedAccounts(ctx context.Context) ([]*Account, error)

	// Existence checks
	ExistsByTeamAndPlatformUser(ctx context.Context, teamID uuid.UUID, platform Platform, platformUserID string) (bool, error)

	// Bulk operations
	BulkUpdateStatus(ctx context.Context, ids []uuid.UUID, status Status) error

	// Cleanup
	HardDelete(ctx context.Context, id uuid.UUID) error
	DeleteExpiredAccounts(ctx context.Context, olderThan time.Duration) (int, error)
}

// TokenRepository handles secure token storage
type TokenRepository interface {
	// Token operations
	SaveTokens(ctx context.Context, accountID uuid.UUID, credentials Credentials) error
	GetTokens(ctx context.Context, accountID uuid.UUID) (*Credentials, error)
	UpdateTokens(ctx context.Context, accountID uuid.UUID, credentials Credentials) error
	DeleteTokens(ctx context.Context, accountID uuid.UUID) error

	// Encryption key management
	RotateEncryptionKey(ctx context.Context, accountID uuid.UUID, newKeyVersion int) error
	GetEncryptionKeyVersion(ctx context.Context, accountID uuid.UUID) (int, error)
}

// WebhookRepository handles webhook events
type WebhookRepository interface {
	// Event storage
	SaveEvent(ctx context.Context, event *WebhookEvent) error
	FindEventByID(ctx context.Context, id uuid.UUID) (*WebhookEvent, error)
	FindEventsByAccount(ctx context.Context, accountID uuid.UUID, limit int) ([]*WebhookEvent, error)
	FindUnprocessedEvents(ctx context.Context, limit int) ([]*WebhookEvent, error)

	// Event processing
	MarkEventProcessed(ctx context.Context, id uuid.UUID, success bool, errorMsg string) error

	// Cleanup
	DeleteOldEvents(ctx context.Context, olderThan time.Duration) (int, error)
}

// PostPublishingRepository tracks post publishing to platforms
type PostPublishingRepository interface {
	// Publishing records
	RecordPublishing(ctx context.Context, record *PublishingRecord) error
	FindByPostID(ctx context.Context, postID uuid.UUID) ([]*PublishingRecord, error)
	FindByAccountID(ctx context.Context, accountID uuid.UUID, limit int) ([]*PublishingRecord, error)
	FindByPlatformPostID(ctx context.Context, platform Platform, platformPostID string) (*PublishingRecord, error)

	// Status updates
	UpdatePublishingStatus(ctx context.Context, id uuid.UUID, status PublishingStatus, error string) error

	// Analytics
	GetPublishingStats(ctx context.Context, accountID uuid.UUID, period time.Duration) (*PublishingStats, error)
}

// PublishingRecord tracks a post published to a platform
type PublishingRecord struct {
	ID             uuid.UUID
	PostID         uuid.UUID
	AccountID      uuid.UUID
	Platform       Platform
	PlatformPostID string
	URL            string
	Status         PublishingStatus
	PublishedAt    time.Time
	Error          string
	RetryCount     int
	Analytics      *PostAnalytics
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// PublishingStatus represents the status of a publishing attempt
type PublishingStatus string

const (
	PublishingStatusPending  PublishingStatus = "pending"
	PublishingStatusSuccess  PublishingStatus = "success"
	PublishingStatusFailed   PublishingStatus = "failed"
	PublishingStatusRetrying PublishingStatus = "retrying"
)

// PublishingStats holds publishing statistics
type PublishingStats struct {
	TotalPublished  int
	SuccessfulPosts int
	FailedPosts     int
	PendingPosts    int
	SuccessRate     float64
	AverageRetries  float64
	TopPlatform     Platform
	ErrorFrequency  map[string]int
}

// AnalyticsRepository handles social media analytics
type AnalyticsRepository interface {
	// Save analytics
	SavePostAnalytics(ctx context.Context, accountID uuid.UUID, postID string, analytics *PostAnalytics) error
	SaveAccountAnalytics(ctx context.Context, accountID uuid.UUID, analytics *AccountAnalytics) error

	// Retrieve analytics
	GetPostAnalytics(ctx context.Context, postID string) (*PostAnalytics, error)
	GetAccountAnalytics(ctx context.Context, accountID uuid.UUID, period time.Duration) (*AccountAnalytics, error)
	GetTeamAnalytics(ctx context.Context, teamID uuid.UUID, period time.Duration) (*TeamSocialAnalytics, error)

	// Aggregations
	GetTopPerformingPosts(ctx context.Context, accountID uuid.UUID, metric string, limit int) ([]*PostAnalytics, error)
	GetEngagementTrends(ctx context.Context, accountID uuid.UUID, days int) ([]*EngagementTrend, error)
}

// TeamSocialAnalytics aggregates analytics across all team accounts
type TeamSocialAnalytics struct {
	TeamID            uuid.UUID
	Period            time.Duration
	TotalAccounts     int
	TotalPosts        int
	TotalImpressions  int
	TotalReach        int
	TotalEngagement   int
	EngagementRate    float64
	PlatformBreakdown map[Platform]*PlatformStats
	TopPerformers     []*PostAnalytics
	UpdatedAt         time.Time
}

// PlatformStats holds platform-specific statistics
type PlatformStats struct {
	Platform       Platform
	Accounts       int
	Posts          int
	Impressions    int
	Reach          int
	Engagement     int
	EngagementRate float64
	GrowthRate     float64
}

// EngagementTrend represents engagement over time
type EngagementTrend struct {
	Date        time.Time
	Impressions int
	Reach       int
	Engagement  int
	Posts       int
	Rate        float64
}

// CacheRepository provides caching for social accounts
type CacheRepository interface {
	// Account caching
	GetAccount(ctx context.Context, id uuid.UUID) (*Account, error)
	SetAccount(ctx context.Context, account *Account, ttl time.Duration) error
	InvalidateAccount(ctx context.Context, id uuid.UUID) error

	// Team accounts caching
	GetTeamAccounts(ctx context.Context, teamID uuid.UUID) ([]*Account, error)
	SetTeamAccounts(ctx context.Context, teamID uuid.UUID, accounts []*Account, ttl time.Duration) error
	InvalidateTeamAccounts(ctx context.Context, teamID uuid.UUID) error

	// Token caching (encrypted)
	GetTokens(ctx context.Context, accountID uuid.UUID) (*Credentials, error)
	SetTokens(ctx context.Context, accountID uuid.UUID, credentials *Credentials, ttl time.Duration) error
	InvalidateTokens(ctx context.Context, accountID uuid.UUID) error
}

// AccountSearchCriteria for complex account queries
type AccountSearchCriteria struct {
	TeamID        *uuid.UUID
	UserID        *uuid.UUID
	Platform      *Platform
	Status        *Status
	AccountType   *AccountType
	HasExpired    *bool
	CreatedAfter  *time.Time
	CreatedBefore *time.Time
}

// AdvancedRepository extends the base repository with complex operations
type AdvancedRepository interface {
	AccountRepository

	// Advanced search
	Search(ctx context.Context, criteria AccountSearchCriteria, offset, limit int) ([]*Account, int64, error)

	// Batch operations
	BatchConnect(ctx context.Context, accounts []*Account) error
	BatchDisconnect(ctx context.Context, accountIDs []uuid.UUID) error

	// Health checks
	GetAccountHealth(ctx context.Context, accountID uuid.UUID) (*AccountHealth, error)
	GetPlatformHealth(ctx context.Context, platform Platform) (*PlatformHealth, error)
}

// AccountHealth represents the health status of an account
type AccountHealth struct {
	AccountID       uuid.UUID
	Status          Status
	TokenValid      bool
	LastSuccessPost time.Time
	FailureRate     float64
	RateLimitStatus string
	RequiresAction  bool
	Issues          []string
}

// PlatformHealth represents the health of a platform integration
type PlatformHealth struct {
	Platform         Platform
	ActiveAccounts   int
	FailedAccounts   int
	RateLimitedCount int
	SuccessRate      float64
	AverageLatency   time.Duration
	LastIncident     *time.Time
	Status           string
}
