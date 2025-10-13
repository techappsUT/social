// ============================================================================
// FILE: backend/cmd/worker/publish_post.go
// PURPOSE: Processor for publishing scheduled posts to social platforms
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

// PublishPostProcessor handles publishing scheduled posts
type PublishPostProcessor struct {
	postRepo     post.Repository
	queueService *services.WorkerQueueService
	logger       common.Logger
	stopChan     chan struct{}
}

// NewPublishPostProcessor creates a new publish post processor
func NewPublishPostProcessor(
	postRepo post.Repository,
	queueService *services.WorkerQueueService,
	logger common.Logger,
) *PublishPostProcessor {
	return &PublishPostProcessor{
		postRepo:     postRepo,
		queueService: queueService,
		logger:       logger,
		stopChan:     make(chan struct{}),
	}
}

// Name returns the processor name
func (p *PublishPostProcessor) Name() string {
	return "PublishPostProcessor"
}

// Run starts the processor loop
func (p *PublishPostProcessor) Run(ctx context.Context) error {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	p.logger.Info("PublishPostProcessor started (polling every 30s)")

	for {
		select {
		case <-ctx.Done():
			p.logger.Info("PublishPostProcessor stopping (context cancelled)")
			return nil
		case <-p.stopChan:
			p.logger.Info("PublishPostProcessor stopped")
			return nil
		case <-ticker.C:
			if err := p.processScheduledPosts(ctx); err != nil {
				p.logger.Error(fmt.Sprintf("Error processing posts: %v", err))
			}
		}
	}
}

// Stop gracefully stops the processor
func (p *PublishPostProcessor) Stop(ctx context.Context) error {
	p.logger.Info("Stopping PublishPostProcessor...")
	close(p.stopChan)
	return nil
}

// processScheduledPosts finds and processes due posts
func (p *PublishPostProcessor) processScheduledPosts(ctx context.Context) error {
	// Find posts that are due for publishing
	duePosts, err := p.postRepo.FindDuePosts(ctx, time.Now())
	if err != nil {
		return fmt.Errorf("failed to find due posts: %w", err)
	}

	if len(duePosts) == 0 {
		return nil // No posts to process
	}

	p.logger.Info(fmt.Sprintf("Found %d posts due for publishing", len(duePosts)))

	// Process each post
	for _, duePost := range duePosts {
		if err := p.publishPost(ctx, duePost); err != nil {
			p.logger.Error(fmt.Sprintf("Failed to publish post %s: %v", duePost.ID(), err))
			continue
		}
	}

	return nil
}

// publishPost publishes a single post
func (p *PublishPostProcessor) publishPost(ctx context.Context, duePost *post.Post) error {
	postID := duePost.ID().String()

	// Try to acquire distributed lock (prevent duplicate processing)
	locked, err := p.queueService.GetQueueLength(ctx, "lock:"+postID)
	if err != nil {
		return fmt.Errorf("failed to check lock: %w", err)
	}
	if locked > 0 {
		p.logger.Warn(fmt.Sprintf("Post %s is already being processed", postID))
		return nil
	}

	// Lock the post for 5 minutes
	payload := map[string]interface{}{
		"post_id": postID,
	}
	if _, err := p.queueService.Enqueue(ctx, "lock:"+postID, payload); err != nil {
		return fmt.Errorf("failed to acquire lock: %w", err)
	}
	defer p.queueService.MarkComplete(ctx, "lock:"+postID, postID)

	// Mark post as publishing
	if err := duePost.MarkPublishing(); err != nil {
		return fmt.Errorf("failed to mark as publishing: %w", err)
	}

	if err := p.postRepo.Update(ctx, duePost); err != nil {
		return fmt.Errorf("failed to update post status: %w", err)
	}

	p.logger.Info(fmt.Sprintf("Publishing post %s to platforms: %v", postID, duePost.Platforms()))

	// TODO: Actually publish to social platforms
	// This would involve:
	// 1. Get social account tokens
	// 2. Call platform adapters (Twitter, LinkedIn, etc.)
	// 3. Handle platform-specific errors
	// 4. Store platform post IDs

	// Simulate publishing (in real implementation, use social adapters)
	time.Sleep(2 * time.Second)

	// For now, mark as published
	now := time.Now()
	if err := duePost.MarkPublished(); err != nil {
		return fmt.Errorf("failed to mark as published: %w", err)
	}

	if err := p.postRepo.Update(ctx, duePost); err != nil {
		return fmt.Errorf("failed to update post: %w", err)
	}

	p.logger.Info(fmt.Sprintf("✅ Successfully published post %s", postID))

	// Enqueue analytics fetch job (fetch metrics after 1 hour)
	analyticsPayload := map[string]interface{}{
		"post_id":    postID,
		"fetch_time": now.Add(1 * time.Hour).Unix(),
	}
	if _, err := p.queueService.Enqueue(ctx, "fetch_analytics", analyticsPayload); err != nil {
		p.logger.Warn(fmt.Sprintf("Failed to enqueue analytics job: %v", err))
	}

	return nil
}

// publishToTwitter publishes content to Twitter (placeholder)
func (p *PublishPostProcessor) publishToTwitter(ctx context.Context, content string) (string, error) {
	// TODO: Implement Twitter API call
	// Use twitter adapter from infrastructure/adapters/social/twitter
	return "twitter_post_id_123", nil
}

// publishToLinkedIn publishes content to LinkedIn (placeholder)
func (p *PublishPostProcessor) publishToLinkedIn(ctx context.Context, content string) (string, error) {
	// TODO: Implement LinkedIn API call
	return "linkedin_post_id_456", nil
}

// publishToFacebook publishes content to Facebook (placeholder)
func (p *PublishPostProcessor) publishToFacebook(ctx context.Context, content string) (string, error) {
	// TODO: Implement Facebook API call
	return "facebook_post_id_789", nil
}
