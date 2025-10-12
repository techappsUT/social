// path: backend/internal/domain/post/service.go
// 🆕 NEW - Clean Architecture

package post

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	teamDomain "github.com/techappsUT/social-queue/internal/domain/team"
)

// Service provides domain-level business logic for posts
type Service struct {
	repo         Repository
	scheduleRepo SchedulerRepository
	teamRepo     teamDomain.Repository
}

// NewService creates a new post domain service
func NewService(repo Repository, scheduleRepo SchedulerRepository, teamRepo teamDomain.Repository) *Service {
	return &Service{
		repo:         repo,
		scheduleRepo: scheduleRepo,
		teamRepo:     teamRepo,
	}
}

// CreatePost creates a new post with team validation
func (s *Service) CreatePost(ctx context.Context, teamID, userID uuid.UUID, content Content, platforms []Platform) (*Post, error) {
	// Verify team exists and is active
	team, err := s.teamRepo.FindByID(ctx, teamID)
	if err != nil {
		return nil, fmt.Errorf("team not found: %w", err)
	}

	if !team.IsActive() {
		return nil, ErrTeamQuotaExceeded
	}

	// Check team's scheduled post limit
	scheduledCount, err := s.repo.CountScheduledByTeam(ctx, teamID)
	if err != nil {
		return nil, fmt.Errorf("failed to count scheduled posts: %w", err)
	}

	if !team.CanSchedulePost(int(scheduledCount)) {
		return nil, ErrScheduleLimitExceeded
	}

	// Create the post
	post, err := NewPost(teamID, userID, content, platforms)
	if err != nil {
		return nil, err
	}

	// Set metadata based on team settings
	if team.Settings().RequireApproval {
		post.metadata.RequiresApproval = true
	}

	// Save post
	if err := s.repo.Create(ctx, post); err != nil {
		return nil, fmt.Errorf("failed to create post: %w", err)
	}

	return post, nil
}

// SchedulePost schedules a post for publishing
func (s *Service) SchedulePost(ctx context.Context, postID uuid.UUID, scheduleTime time.Time, scheduledBy uuid.UUID) error {
	// Get post
	post, err := s.repo.FindByID(ctx, postID)
	if err != nil {
		return ErrPostNotFound
	}

	// Verify scheduler has permission (would check via team member repo in real impl)
	// For now, we'll assume the check is done at the use case level

	// Check for scheduling conflicts (optional, based on team settings)
	conflicts, err := s.repo.FindScheduledBetween(
		ctx,
		post.teamID,
		scheduleTime.Add(-5*time.Minute),
		scheduleTime.Add(5*time.Minute),
	)
	if err == nil && len(conflicts) > 0 {
		// Check if team allows concurrent posts
		team, _ := s.teamRepo.FindByID(ctx, post.teamID)
		if team != nil && !team.Settings().AutoSchedule {
			return ErrScheduleConflict
		}
	}

	// Schedule the post
	if err := post.Schedule(scheduleTime); err != nil {
		return err
	}

	// Update in repository
	if err := s.repo.Update(ctx, post); err != nil {
		return fmt.Errorf("failed to schedule post: %w", err)
	}

	// Add to scheduler queue if time is near
	if scheduleTime.Before(time.Now().Add(24 * time.Hour)) {
		if err := s.scheduleRepo.AddToQueue(ctx, post.ID(), post.Priority()); err != nil {
			// Log error but don't fail - queue will pick it up later
		}
	}

	return nil
}

// ApprovePost approves a post for publishing
func (s *Service) ApprovePost(ctx context.Context, postID, approverID uuid.UUID) error {
	// Get post
	post, err := s.repo.FindByID(ctx, postID)
	if err != nil {
		return ErrPostNotFound
	}

	// Check if approver created the post
	if post.CreatedBy() == approverID {
		return ErrCannotApproveOwnPost
	}

	// Approve the post
	if err := post.Approve(approverID); err != nil {
		return err
	}

	// Update in repository
	if err := s.repo.Update(ctx, post); err != nil {
		return fmt.Errorf("failed to approve post: %w", err)
	}

	// If post is due, add to queue
	if post.IsDue() {
		if err := post.Queue(); err == nil {
			s.repo.Update(ctx, post)
			s.scheduleRepo.AddToQueue(ctx, post.ID(), post.Priority())
		}
	}

	return nil
}

// ProcessDuePosts processes posts that are due for publishing
func (s *Service) ProcessDuePosts(ctx context.Context) (int, error) {
	// Find all due posts
	duePosts, err := s.repo.FindDuePosts(ctx, time.Now())
	if err != nil {
		return 0, fmt.Errorf("failed to find due posts: %w", err)
	}

	processed := 0
	for _, post := range duePosts {
		// Check approval if required
		if post.NeedsApproval() {
			continue
		}

		// Queue the post
		if err := post.Queue(); err != nil {
			continue
		}

		// Update status
		if err := s.repo.Update(ctx, post); err != nil {
			continue
		}

		// Add to processing queue
		if err := s.scheduleRepo.AddToQueue(ctx, post.ID(), post.Priority()); err != nil {
			continue
		}

		processed++
	}

	return processed, nil
}

// PublishPost marks a post as being published
func (s *Service) PublishPost(ctx context.Context, postID uuid.UUID) error {
	// Lock post for processing
	if err := s.repo.LockForProcessing(ctx, postID); err != nil {
		return fmt.Errorf("failed to lock post: %w", err)
	}
	defer s.repo.UnlockPost(ctx, postID)

	// Get post
	post, err := s.repo.FindByID(ctx, postID)
	if err != nil {
		return ErrPostNotFound
	}

	// Verify post can be published
	if !post.CanPublish() {
		return ErrNotQueued
	}

	// Check daily limit for team
	publishedToday, err := s.repo.CountPublishedToday(ctx, post.TeamID())
	if err != nil {
		return fmt.Errorf("failed to check daily limit: %w", err)
	}

	team, err := s.teamRepo.FindByID(ctx, post.TeamID())
	if err != nil {
		return fmt.Errorf("team not found: %w", err)
	}

	if !team.CanPublishToday(int(publishedToday)) {
		return ErrDailyLimitExceeded
	}

	// Mark as publishing
	if err := post.MarkPublishing(); err != nil {
		return err
	}

	// Update in repository
	if err := s.repo.Update(ctx, post); err != nil {
		return fmt.Errorf("failed to update post status: %w", err)
	}

	// Note: Actual publishing to platforms would be handled by
	// infrastructure layer (social adapters) called from use case

	return nil
}

// MarkPostPublished marks a post as successfully published
func (s *Service) MarkPostPublished(ctx context.Context, postID uuid.UUID) error {
	post, err := s.repo.FindByID(ctx, postID)
	if err != nil {
		return ErrPostNotFound
	}

	if err := post.MarkPublished(); err != nil {
		return err
	}

	if err := s.repo.Update(ctx, post); err != nil {
		return fmt.Errorf("failed to mark post as published: %w", err)
	}

	// Remove from queue
	s.scheduleRepo.RemoveFromQueue(ctx, postID)

	return nil
}

// MarkPostFailed marks a post as failed with retry logic
func (s *Service) MarkPostFailed(ctx context.Context, postID uuid.UUID, errorMessage string) error {
	post, err := s.repo.FindByID(ctx, postID)
	if err != nil {
		return ErrPostNotFound
	}

	if err := post.MarkFailed(errorMessage); err != nil {
		return err
	}

	// Check if can retry
	if post.CanRetry() {
		// Reschedule for retry (exponential backoff)
		retryTime := time.Now().Add(time.Duration(post.metadata.RetryCount) * 10 * time.Minute)
		post.scheduleTime = &retryTime
		post.status = StatusScheduled
	}

	if err := s.repo.Update(ctx, post); err != nil {
		return fmt.Errorf("failed to mark post as failed: %w", err)
	}

	return nil
}

// CancelPost cancels a scheduled post
func (s *Service) CancelPost(ctx context.Context, postID uuid.UUID, canceledBy uuid.UUID) error {
	post, err := s.repo.FindByID(ctx, postID)
	if err != nil {
		return ErrPostNotFound
	}

	// Verify canceler has permission (would check via team member repo)

	if err := post.Cancel(); err != nil {
		return err
	}

	if err := s.repo.Update(ctx, post); err != nil {
		return fmt.Errorf("failed to cancel post: %w", err)
	}

	// Remove from queue if present
	s.scheduleRepo.RemoveFromQueue(ctx, postID)

	return nil
}

// UpdatePostAnalytics updates analytics for a published post
func (s *Service) UpdatePostAnalytics(ctx context.Context, postID uuid.UUID, analytics Analytics) error {
	post, err := s.repo.FindByID(ctx, postID)
	if err != nil {
		return ErrPostNotFound
	}

	if post.Status() != StatusPublished {
		return ErrAnalyticsNotAvailable
	}

	post.UpdateAnalytics(analytics)

	if err := s.repo.Update(ctx, post); err != nil {
		return fmt.Errorf("failed to update analytics: %w", err)
	}

	return nil
}

// ValidatePostForPlatforms validates post content for all selected platforms
func (s *Service) ValidatePostForPlatforms(ctx context.Context, postID uuid.UUID) error {
	post, err := s.repo.FindByID(ctx, postID)
	if err != nil {
		return ErrPostNotFound
	}

	// Validate for each platform
	for _, platform := range post.Platforms() {
		if err := post.ValidateForPlatform(platform); err != nil {
			return fmt.Errorf("validation failed for %s: %w", platform, err)
		}
	}

	return nil
}

// CleanupOldDrafts removes old draft posts
func (s *Service) CleanupOldDrafts(ctx context.Context, olderThan time.Duration) (int, error) {
	cutoff := time.Now().Add(-olderThan)
	deleted, err := s.repo.DeleteOldDrafts(ctx, cutoff)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup old drafts: %w", err)
	}
	return deleted, nil
}

// GetQueueStatus gets the current queue status
func (s *Service) GetQueueStatus(ctx context.Context) (map[string]interface{}, error) {
	queueSize, err := s.scheduleRepo.GetQueueSize(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get queue size: %w", err)
	}

	nextBatch, err := s.scheduleRepo.GetNextScheduledBatch(ctx, 10)
	if err != nil {
		return nil, fmt.Errorf("failed to get next batch: %w", err)
	}

	status := map[string]interface{}{
		"queue_size":   queueSize,
		"next_batch":   len(nextBatch),
		"next_post_at": nil,
	}

	if len(nextBatch) > 0 && nextBatch[0].ScheduleTime() != nil {
		status["next_post_at"] = nextBatch[0].ScheduleTime()
	}

	return status, nil
}

// Specifications for complex business rules

// PublishablePostSpecification checks if a post can be published
type PublishablePostSpecification struct{}

func (s PublishablePostSpecification) IsSatisfiedBy(post *Post) bool {
	return post.CanPublish()
}

// ApprovablePostSpecification checks if a post needs approval
type ApprovablePostSpecification struct{}

func (s ApprovablePostSpecification) IsSatisfiedBy(post *Post) bool {
	return post.NeedsApproval()
}

// RetryablePostSpecification checks if a post can be retried
type RetryablePostSpecification struct{}

func (s RetryablePostSpecification) IsSatisfiedBy(post *Post) bool {
	return post.CanRetry()
}
