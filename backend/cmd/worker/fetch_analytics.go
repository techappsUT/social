// ============================================================================
// FILE: backend/cmd/worker/fetch_analytics.go
// PURPOSE: Processor for fetching analytics from social platforms
// ============================================================================

package main

import (
	"context"
	"fmt"
	"time"

	"github.com/techappsUT/social-queue/internal/application/common"
	"github.com/techappsUT/social-queue/internal/domain/post"
	"github.com/techappsUT/social-queue/internal/infrastructure/services"
)

// FetchAnalyticsProcessor handles fetching analytics from social platforms
type FetchAnalyticsProcessor struct {
	postRepo     post.Repository
	queueService *services.WorkerQueueService
	logger       common.Logger
	stopChan     chan struct{}
}

// NewFetchAnalyticsProcessor creates a new analytics processor
func NewFetchAnalyticsProcessor(
	postRepo post.Repository,
	queueService *services.WorkerQueueService,
	logger common.Logger,
) *FetchAnalyticsProcessor {
	return &FetchAnalyticsProcessor{
		postRepo:     postRepo,
		queueService: queueService,
		logger:       logger,
		stopChan:     make(chan struct{}),
	}
}

// Name returns the processor name
func (p *FetchAnalyticsProcessor) Name() string {
	return "FetchAnalyticsProcessor"
}

// Run starts the processor loop
func (p *FetchAnalyticsProcessor) Run(ctx context.Context) error {
	ticker := time.NewTicker(6 * time.Hour)
	defer ticker.Stop()

	p.logger.Info("FetchAnalyticsProcessor started (runs every 6 hours)")

	// Run immediately on startup
	if err := p.fetchAnalytics(ctx); err != nil {
		p.logger.Error(fmt.Sprintf("Error fetching analytics: %v", err))
	}

	for {
		select {
		case <-ctx.Done():
			p.logger.Info("FetchAnalyticsProcessor stopping (context cancelled)")
			return nil
		case <-p.stopChan:
			p.logger.Info("FetchAnalyticsProcessor stopped")
			return nil
		case <-ticker.C:
			if err := p.fetchAnalytics(ctx); err != nil {
				p.logger.Error(fmt.Sprintf("Error fetching analytics: %v", err))
			}
		}
	}
}

// Stop gracefully stops the processor
func (p *FetchAnalyticsProcessor) Stop(ctx context.Context) error {
	p.logger.Info("Stopping FetchAnalyticsProcessor...")
	close(p.stopChan)
	return nil
}

// fetchAnalytics fetches analytics for published posts
func (p *FetchAnalyticsProcessor) fetchAnalytics(ctx context.Context) error {
	p.logger.Info("Starting analytics fetch...")

	// Find published posts that need analytics update
	// (typically posts published in last 30 days)
	// Fetch up to 100 published posts at a time
	offset := 0
	limit := 100
	posts, err := p.postRepo.FindByStatus(ctx, post.StatusPublished, offset, limit)
	if err != nil {
		return fmt.Errorf("failed to find published posts: %w", err)
	}

	if len(posts) == 0 {
		p.logger.Info("No published posts to fetch analytics for")
		return nil
	}

	p.logger.Info(fmt.Sprintf("Found %d published posts to fetch analytics for", len(posts)))

	successCount := 0
	failureCount := 0

	for _, post := range posts {
		if err := p.fetchPostAnalytics(ctx, post); err != nil {
			p.logger.Error(fmt.Sprintf("Failed to fetch analytics for post %s: %v", post.ID(), err))
			failureCount++
			continue
		}
		successCount++
	}

	p.logger.Info(fmt.Sprintf("✅ Analytics fetch completed: %d succeeded, %d failed", successCount, failureCount))
	return nil
}

// fetchPostAnalytics fetches analytics for a single post
func (p *FetchAnalyticsProcessor) fetchPostAnalytics(ctx context.Context, postEntity *post.Post) error {
	postID := postEntity.ID()

	// TODO: Call social platform API to fetch real metrics
	// For now, we'll simulate with placeholder data

	// ✅ FIXED: Use correct Analytics struct fields
	analytics := post.Analytics{
		Impressions: 1000,
		Clicks:      50,
		Likes:       25,
		Shares:      5,
		Comments:    3,
		Reach:       800,
		Engagement:  0.033,            // (25 + 5 + 3) / 1000 = 0.033 or 3.3%
		LastUpdated: time.Now().UTC(), // ✅ FIXED: Use LastUpdated instead of FetchedAt
	}

	// Update post with new analytics
	postEntity.UpdateAnalytics(analytics)

	// Save to database
	if err := p.postRepo.Update(ctx, postEntity); err != nil {
		return fmt.Errorf("failed to save analytics: %w", err)
	}

	// ✅ FIXED: Calculate total engagement from individual metrics
	totalEngagement := analytics.Likes + analytics.Shares + analytics.Comments
	p.logger.Info(fmt.Sprintf("✓ Updated analytics for post %s: %d impressions, %d total engagements",
		postID, analytics.Impressions, totalEngagement))

	return nil
}

// Helper function to calculate engagement rate
func calculateEngagementRate(impressions, likes, shares, comments int) float64 {
	if impressions == 0 {
		return 0.0
	}
	totalEngagement := likes + shares + comments
	return float64(totalEngagement) / float64(impressions)
}
