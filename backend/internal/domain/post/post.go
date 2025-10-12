// path: backend/internal/domain/post/post.go
// 🆕 NEW - Clean Architecture

package post

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Post represents a social media post to be published
type Post struct {
	id           uuid.UUID
	teamID       uuid.UUID
	createdBy    uuid.UUID
	content      Content
	platforms    []Platform
	scheduleTime *time.Time
	publishedAt  *time.Time
	status       Status
	priority     Priority
	metadata     Metadata
	analytics    *Analytics
	createdAt    time.Time
	updatedAt    time.Time
	deletedAt    *time.Time
}

// Content holds the post content
type Content struct {
	Text        string
	MediaURLs   []string
	MediaTypes  []MediaType
	Hashtags    []string
	Mentions    []string
	Link        string
	LinkPreview *LinkPreview
}

// MediaType represents the type of media
type MediaType string

const (
	MediaTypeImage MediaType = "image"
	MediaTypeVideo MediaType = "video"
	MediaTypeGIF   MediaType = "gif"
)

// LinkPreview holds Open Graph data for links
type LinkPreview struct {
	Title       string
	Description string
	ImageURL    string
	SiteName    string
}

// Platform represents a social media platform
type Platform string

const (
	PlatformTwitter   Platform = "twitter"
	PlatformFacebook  Platform = "facebook"
	PlatformLinkedIn  Platform = "linkedin"
	PlatformInstagram Platform = "instagram"
	PlatformTikTok    Platform = "tiktok"
	PlatformPinterest Platform = "pinterest"
)

// Status represents the post status
type Status string

const (
	StatusDraft      Status = "draft"
	StatusScheduled  Status = "scheduled"
	StatusQueued     Status = "queued"
	StatusPublishing Status = "publishing"
	StatusPublished  Status = "published"
	StatusFailed     Status = "failed"
	StatusCanceled   Status = "canceled"
)

// Priority represents post priority in queue
type Priority int

const (
	PriorityLow    Priority = 0
	PriorityNormal Priority = 1
	PriorityHigh   Priority = 2
	PriorityUrgent Priority = 3
)

// Metadata holds additional post information
type Metadata struct {
	Campaign         string
	Tags             []string
	InternalNote     string
	ApprovedBy       *uuid.UUID
	ApprovedAt       *time.Time
	RequiresApproval bool
	RetryCount       int
	LastError        string
	CustomFields     map[string]interface{}
}

// Analytics holds post performance metrics
type Analytics struct {
	Impressions int
	Clicks      int
	Likes       int
	Shares      int
	Comments    int
	Reach       int
	Engagement  float64
	LastUpdated time.Time
}

// NewPost creates a new post with validation
func NewPost(teamID, createdBy uuid.UUID, content Content, platforms []Platform) (*Post, error) {
	// Validate IDs
	if teamID == uuid.Nil {
		return nil, ErrInvalidTeamID
	}
	if createdBy == uuid.Nil {
		return nil, ErrInvalidUserID
	}

	// Validate content
	if err := validateContent(content); err != nil {
		return nil, err
	}

	// Validate platforms
	if len(platforms) == 0 {
		return nil, ErrNoPlatformsSelected
	}
	for _, platform := range platforms {
		if !isValidPlatform(platform) {
			return nil, ErrInvalidPlatform
		}
	}

	now := time.Now().UTC()

	return &Post{
		id:        uuid.New(),
		teamID:    teamID,
		createdBy: createdBy,
		content:   content,
		platforms: platforms,
		status:    StatusDraft,
		priority:  PriorityNormal,
		metadata: Metadata{
			Tags:         []string{},
			RetryCount:   0,
			CustomFields: make(map[string]interface{}),
		},
		createdAt: now,
		updatedAt: now,
	}, nil
}

// Reconstruct recreates a post from persistence
func Reconstruct(
	id uuid.UUID,
	teamID uuid.UUID,
	createdBy uuid.UUID,
	content Content,
	platforms []Platform,
	scheduleTime *time.Time,
	publishedAt *time.Time,
	status Status,
	priority Priority,
	metadata Metadata,
	analytics *Analytics,
	createdAt time.Time,
	updatedAt time.Time,
	deletedAt *time.Time,
) *Post {
	return &Post{
		id:           id,
		teamID:       teamID,
		createdBy:    createdBy,
		content:      content,
		platforms:    platforms,
		scheduleTime: scheduleTime,
		publishedAt:  publishedAt,
		status:       status,
		priority:     priority,
		metadata:     metadata,
		analytics:    analytics,
		createdAt:    createdAt,
		updatedAt:    updatedAt,
		deletedAt:    deletedAt,
	}
}

// Getters
func (p *Post) ID() uuid.UUID            { return p.id }
func (p *Post) TeamID() uuid.UUID        { return p.teamID }
func (p *Post) CreatedBy() uuid.UUID     { return p.createdBy }
func (p *Post) Content() Content         { return p.content }
func (p *Post) Platforms() []Platform    { return p.platforms }
func (p *Post) ScheduleTime() *time.Time { return p.scheduleTime }
func (p *Post) PublishedAt() *time.Time  { return p.publishedAt }
func (p *Post) Status() Status           { return p.status }
func (p *Post) Priority() Priority       { return p.priority }
func (p *Post) Metadata() Metadata       { return p.metadata }
func (p *Post) Analytics() *Analytics    { return p.analytics }
func (p *Post) CreatedAt() time.Time     { return p.createdAt }
func (p *Post) UpdatedAt() time.Time     { return p.updatedAt }
func (p *Post) DeletedAt() *time.Time    { return p.deletedAt }

// Business Logic Methods

// Schedule schedules the post for a specific time
func (p *Post) Schedule(scheduleTime time.Time) error {
	if p.status == StatusPublished {
		return ErrCannotSchedulePublished
	}

	if p.status == StatusCanceled {
		return ErrPostCanceled
	}

	if scheduleTime.Before(time.Now()) {
		return ErrScheduleTimeInPast
	}

	// Max 1 year in future
	if scheduleTime.After(time.Now().AddDate(1, 0, 0)) {
		return ErrScheduleTimeTooFar
	}

	p.scheduleTime = &scheduleTime
	p.status = StatusScheduled
	p.updatedAt = time.Now().UTC()
	return nil
}

// UpdateContent updates the post content
func (p *Post) UpdateContent(content Content) error {
	if p.status == StatusPublished {
		return ErrCannotEditPublished
	}

	if p.status == StatusPublishing {
		return ErrCannotEditWhilePublishing
	}

	if err := validateContent(content); err != nil {
		return err
	}

	p.content = content
	p.updatedAt = time.Now().UTC()

	// Reset approval if required
	if p.metadata.RequiresApproval && p.metadata.ApprovedBy != nil {
		p.metadata.ApprovedBy = nil
		p.metadata.ApprovedAt = nil
	}

	return nil
}

// UpdatePlatforms updates target platforms
func (p *Post) UpdatePlatforms(platforms []Platform) error {
	if p.status == StatusPublished || p.status == StatusPublishing {
		return ErrCannotEditPublished
	}

	if len(platforms) == 0 {
		return ErrNoPlatformsSelected
	}

	for _, platform := range platforms {
		if !isValidPlatform(platform) {
			return ErrInvalidPlatform
		}
	}

	p.platforms = platforms
	p.updatedAt = time.Now().UTC()
	return nil
}

// Approve approves the post for publishing
func (p *Post) Approve(approverID uuid.UUID) error {
	if !p.metadata.RequiresApproval {
		return ErrApprovalNotRequired
	}

	if p.metadata.ApprovedBy != nil {
		return ErrAlreadyApproved
	}

	now := time.Now().UTC()
	p.metadata.ApprovedBy = &approverID
	p.metadata.ApprovedAt = &now
	p.updatedAt = now
	return nil
}

// Queue adds the post to publishing queue
func (p *Post) Queue() error {
	if p.status != StatusScheduled {
		return ErrNotScheduled
	}

	if p.metadata.RequiresApproval && p.metadata.ApprovedBy == nil {
		return ErrRequiresApproval
	}

	p.status = StatusQueued
	p.updatedAt = time.Now().UTC()
	return nil
}

// MarkPublishing marks the post as being published
func (p *Post) MarkPublishing() error {
	if p.status != StatusQueued {
		return ErrNotQueued
	}

	p.status = StatusPublishing
	p.updatedAt = time.Now().UTC()
	return nil
}

// MarkPublished marks the post as successfully published
func (p *Post) MarkPublished() error {
	if p.status != StatusPublishing {
		return ErrNotPublishing
	}

	now := time.Now().UTC()
	p.status = StatusPublished
	p.publishedAt = &now
	p.updatedAt = now
	return nil
}

// MarkFailed marks the post as failed to publish
func (p *Post) MarkFailed(errorMessage string) error {
	p.status = StatusFailed
	p.metadata.LastError = errorMessage
	p.metadata.RetryCount++
	p.updatedAt = time.Now().UTC()
	return nil
}

// Cancel cancels a scheduled post
func (p *Post) Cancel() error {
	if p.status == StatusPublished {
		return ErrCannotCancelPublished
	}

	if p.status == StatusCanceled {
		return ErrAlreadyCanceled
	}

	p.status = StatusCanceled
	p.updatedAt = time.Now().UTC()
	return nil
}

// SoftDelete soft deletes the post
func (p *Post) SoftDelete() error {
	if p.deletedAt != nil {
		return ErrPostAlreadyDeleted
	}

	now := time.Now().UTC()
	p.deletedAt = &now
	p.updatedAt = now
	return nil
}

// Restore restores a soft-deleted post
func (p *Post) Restore() error {
	if p.deletedAt == nil {
		return ErrPostNotDeleted
	}

	p.deletedAt = nil
	p.updatedAt = time.Now().UTC()
	return nil
}

// SetPriority sets the post priority
func (p *Post) SetPriority(priority Priority) error {
	if !isValidPriority(priority) {
		return ErrInvalidPriority
	}

	p.priority = priority
	p.updatedAt = time.Now().UTC()
	return nil
}

// UpdateAnalytics updates post analytics
func (p *Post) UpdateAnalytics(analytics Analytics) {
	analytics.LastUpdated = time.Now().UTC()
	p.analytics = &analytics
	p.updatedAt = time.Now().UTC()
}

// Business Rule Checks

// CanPublish checks if the post can be published
func (p *Post) CanPublish() bool {
	if p.status != StatusQueued {
		return false
	}

	if p.metadata.RequiresApproval && p.metadata.ApprovedBy == nil {
		return false
	}

	if p.scheduleTime != nil && p.scheduleTime.After(time.Now()) {
		return false
	}

	return true
}

// IsScheduled checks if the post is scheduled
func (p *Post) IsScheduled() bool {
	return p.status == StatusScheduled && p.scheduleTime != nil
}

// IsDue checks if a scheduled post is due for publishing
func (p *Post) IsDue() bool {
	if !p.IsScheduled() {
		return false
	}

	return p.scheduleTime.Before(time.Now()) || p.scheduleTime.Equal(time.Now())
}

// CanRetry checks if a failed post can be retried
func (p *Post) CanRetry() bool {
	return p.status == StatusFailed && p.metadata.RetryCount < 3
}

// NeedsApproval checks if the post needs approval
func (p *Post) NeedsApproval() bool {
	return p.metadata.RequiresApproval && p.metadata.ApprovedBy == nil
}

// GetCharacterCount returns character count for different platforms
func (p *Post) GetCharacterCount() map[Platform]int {
	counts := make(map[Platform]int)
	baseText := p.content.Text

	for _, platform := range p.platforms {
		switch platform {
		case PlatformTwitter:
			// Twitter counts URLs as 23 chars
			count := len([]rune(baseText))
			if p.content.Link != "" {
				count += 23
			}
			counts[platform] = count
		case PlatformFacebook:
			counts[platform] = len([]rune(baseText))
		case PlatformLinkedIn:
			counts[platform] = len([]rune(baseText))
		default:
			counts[platform] = len([]rune(baseText))
		}
	}

	return counts
}

// ValidateForPlatform validates content for specific platform
func (p *Post) ValidateForPlatform(platform Platform) error {
	switch platform {
	case PlatformTwitter:
		if p.GetCharacterCount()[platform] > 280 {
			return ErrContentTooLongForPlatform
		}
		if len(p.content.MediaURLs) > 4 {
			return ErrTooManyMediaFiles
		}
	case PlatformInstagram:
		if len(p.content.MediaURLs) == 0 {
			return ErrInstagramRequiresMedia
		}
		if len(p.content.MediaURLs) > 10 {
			return ErrTooManyMediaFiles
		}
	case PlatformLinkedIn:
		if len([]rune(p.content.Text)) > 3000 {
			return ErrContentTooLongForPlatform
		}
	}
	return nil
}

// Helper Functions

func validateContent(content Content) error {
	// Check if content is empty
	if strings.TrimSpace(content.Text) == "" && len(content.MediaURLs) == 0 {
		return ErrEmptyContent
	}

	// Validate media count
	if len(content.MediaURLs) > 10 {
		return ErrTooManyMediaFiles
	}

	// Validate media types match URLs
	if len(content.MediaURLs) != len(content.MediaTypes) {
		return ErrMediaTypeMismatch
	}

	return nil
}

func isValidPlatform(platform Platform) bool {
	switch platform {
	case PlatformTwitter, PlatformFacebook, PlatformLinkedIn,
		PlatformInstagram, PlatformTikTok, PlatformPinterest:
		return true
	default:
		return false
	}
}

func isValidPriority(priority Priority) bool {
	return priority >= PriorityLow && priority <= PriorityUrgent
}

// JSON serialization for storing complex fields in database

func (c Content) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Text        string       `json:"text"`
		MediaURLs   []string     `json:"media_urls"`
		MediaTypes  []MediaType  `json:"media_types"`
		Hashtags    []string     `json:"hashtags"`
		Mentions    []string     `json:"mentions"`
		Link        string       `json:"link"`
		LinkPreview *LinkPreview `json:"link_preview,omitempty"`
	}{
		Text:        c.Text,
		MediaURLs:   c.MediaURLs,
		MediaTypes:  c.MediaTypes,
		Hashtags:    c.Hashtags,
		Mentions:    c.Mentions,
		Link:        c.Link,
		LinkPreview: c.LinkPreview,
	})
}
