// path: backend/internal/domain/post/repository.go
// 🆕 NEW - Clean Architecture

package post

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Repository defines the interface for post persistence
type Repository interface {
	// Basic CRUD operations
	Create(ctx context.Context, post *Post) error
	Update(ctx context.Context, post *Post) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (*Post, error)

	// Team-based queries
	FindByTeamID(ctx context.Context, teamID uuid.UUID, offset, limit int) ([]*Post, error)
	FindByUserID(ctx context.Context, userID uuid.UUID, offset, limit int) ([]*Post, error)
	CountByTeamID(ctx context.Context, teamID uuid.UUID) (int64, error)

	// Status-based queries
	FindByStatus(ctx context.Context, status Status, offset, limit int) ([]*Post, error)
	FindScheduled(ctx context.Context, offset, limit int) ([]*Post, error)
	FindDuePosts(ctx context.Context, before time.Time) ([]*Post, error)
	FindQueued(ctx context.Context, limit int) ([]*Post, error)
	FindPublished(ctx context.Context, teamID uuid.UUID, offset, limit int) ([]*Post, error)
	FindFailed(ctx context.Context, limit int) ([]*Post, error)

	// Platform-specific queries
	FindByPlatform(ctx context.Context, platform Platform, offset, limit int) ([]*Post, error)
	FindByTeamAndPlatform(ctx context.Context, teamID uuid.UUID, platform Platform, offset, limit int) ([]*Post, error)

	// Time-based queries
	FindScheduledBetween(ctx context.Context, teamID uuid.UUID, start, end time.Time) ([]*Post, error)
	FindPublishedBetween(ctx context.Context, teamID uuid.UUID, start, end time.Time) ([]*Post, error)
	FindCreatedToday(ctx context.Context, teamID uuid.UUID) ([]*Post, error)

	// Analytics queries
	CountScheduledByTeam(ctx context.Context, teamID uuid.UUID) (int64, error)
	CountPublishedToday(ctx context.Context, teamID uuid.UUID) (int64, error)
	GetTeamPostStats(ctx context.Context, teamID uuid.UUID) (*PostStatistics, error)

	// Bulk operations
	BulkUpdateStatus(ctx context.Context, ids []uuid.UUID, status Status) error
	MarkOverduePosts(ctx context.Context, before time.Time) (int, error)

	// Queue management
	GetNextInQueue(ctx context.Context) (*Post, error)
	LockForProcessing(ctx context.Context, id uuid.UUID) error
	UnlockPost(ctx context.Context, id uuid.UUID) error

	// Cleanup operations
	HardDelete(ctx context.Context, id uuid.UUID) error
	DeleteOldDrafts(ctx context.Context, olderThan time.Time) (int, error)
	Restore(ctx context.Context, id uuid.UUID) error
}

// SchedulerRepository handles scheduling-specific operations
type SchedulerRepository interface {
	// Queue management
	AddToQueue(ctx context.Context, postID uuid.UUID, priority Priority) error
	RemoveFromQueue(ctx context.Context, postID uuid.UUID) error
	GetQueuePosition(ctx context.Context, postID uuid.UUID) (int, error)
	GetQueueSize(ctx context.Context) (int, error)

	// Scheduling operations
	GetNextScheduledBatch(ctx context.Context, limit int) ([]*Post, error)
	ReschedulePost(ctx context.Context, postID uuid.UUID, newTime time.Time) error

	// Rate limiting
	GetPublishCountForAccount(ctx context.Context, accountID uuid.UUID, since time.Time) (int, error)
	CanPublishToAccount(ctx context.Context, accountID uuid.UUID) (bool, error)
}

// AnalyticsRepository handles post analytics
type AnalyticsRepository interface {
	// Save analytics
	SaveAnalytics(ctx context.Context, postID uuid.UUID, analytics *Analytics) error
	UpdateAnalytics(ctx context.Context, postID uuid.UUID, analytics *Analytics) error

	// Retrieve analytics
	GetAnalytics(ctx context.Context, postID uuid.UUID) (*Analytics, error)
	GetTeamAnalytics(ctx context.Context, teamID uuid.UUID, period time.Duration) (*TeamAnalytics, error)
	GetPlatformAnalytics(ctx context.Context, teamID uuid.UUID, platform Platform, period time.Duration) (*PlatformAnalytics, error)

	// Aggregations
	GetTopPosts(ctx context.Context, teamID uuid.UUID, metric string, limit int) ([]*Post, error)
	GetEngagementRate(ctx context.Context, teamID uuid.UUID, period time.Duration) (float64, error)
}

// PostStatistics holds aggregated post statistics
type PostStatistics struct {
	TotalPosts     int64
	ScheduledPosts int64
	PublishedPosts int64
	FailedPosts    int64
	DraftPosts     int64
	PostsToday     int64
	PostsThisWeek  int64
	PostsThisMonth int64
	AveragePerDay  float64
	TopPlatform    Platform
	TopPublishHour int
}

// TeamAnalytics holds team-level analytics
type TeamAnalytics struct {
	TeamID             uuid.UUID
	Period             time.Duration
	TotalPosts         int
	TotalImpressions   int
	TotalClicks        int
	TotalEngagement    int
	EngagementRate     float64
	BestPerformingPost *Post
	PlatformBreakdown  map[Platform]*PlatformAnalytics
}

// PlatformAnalytics holds platform-specific analytics
type PlatformAnalytics struct {
	Platform       Platform
	Posts          int
	Impressions    int
	Clicks         int
	Likes          int
	Shares         int
	Comments       int
	EngagementRate float64
	BestTime       time.Time
}

// CacheRepository defines caching operations for posts
type CacheRepository interface {
	Get(ctx context.Context, id uuid.UUID) (*Post, error)
	Set(ctx context.Context, post *Post, ttl time.Duration) error
	Delete(ctx context.Context, id uuid.UUID) error
	InvalidateTeamPosts(ctx context.Context, teamID uuid.UUID) error

	// Queue cache
	GetQueue(ctx context.Context) ([]*Post, error)
	SetQueue(ctx context.Context, posts []*Post, ttl time.Duration) error
	InvalidateQueue(ctx context.Context) error
}

// SearchCriteria for complex post queries
type SearchCriteria struct {
	Query            string
	TeamID           *uuid.UUID
	UserID           *uuid.UUID
	Status           *Status
	Platform         *Platform
	Priority         *Priority
	HasMedia         *bool
	Campaign         *string
	Tags             []string
	ScheduledAfter   *time.Time
	ScheduledBefore  *time.Time
	PublishedAfter   *time.Time
	PublishedBefore  *time.Time
	RequiresApproval *bool
	IsApproved       *bool
}

// QueryOptions for pagination and sorting
type QueryOptions struct {
	Offset           int
	Limit            int
	SortBy           string // "created_at", "scheduled_time", "priority", etc.
	SortOrder        string // "asc" or "desc"
	IncludeDeleted   bool
	IncludeAnalytics bool
}

// AdvancedRepository extends Repository with complex operations
type AdvancedRepository interface {
	Repository

	// Advanced search
	Search(ctx context.Context, criteria SearchCriteria, opts QueryOptions) ([]*Post, int64, error)

	// Bulk operations
	BulkCreate(ctx context.Context, posts []*Post) error
	BulkSchedule(ctx context.Context, postIDs []uuid.UUID, scheduleTime time.Time) error
	BulkCancel(ctx context.Context, postIDs []uuid.UUID) error

	// Campaign operations
	FindByCampaign(ctx context.Context, teamID uuid.UUID, campaign string) ([]*Post, error)
	GetCampaignStats(ctx context.Context, teamID uuid.UUID, campaign string) (*CampaignStats, error)

	// Duplicate detection
	FindDuplicates(ctx context.Context, teamID uuid.UUID, content string, timeWindow time.Duration) ([]*Post, error)
}

// CampaignStats holds campaign-level statistics
type CampaignStats struct {
	Campaign        string
	TotalPosts      int
	PublishedPosts  int
	TotalReach      int
	TotalEngagement int
	TopPerformer    *Post
	Platforms       []Platform
}

// Transaction represents a database transaction
type Transaction interface {
	Commit() error
	Rollback() error
}
