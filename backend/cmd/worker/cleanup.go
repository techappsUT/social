// ============================================================================
// FILE: backend/cmd/worker/cleanup.go
// PURPOSE: Processor for database cleanup and maintenance tasks
// ============================================================================

package main

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/techappsUT/social-queue/internal/application/common"
	"github.com/techappsUT/social-queue/internal/infrastructure/services"
)

// CleanupProcessor handles database cleanup and maintenance
type CleanupProcessor struct {
	db           *sql.DB
	queueService *services.WorkerQueueService
	logger       common.Logger
	stopChan     chan struct{}
}

// NewCleanupProcessor creates a new cleanup processor
func NewCleanupProcessor(
	db *sql.DB,
	queueService *services.WorkerQueueService,
	logger common.Logger,
) *CleanupProcessor {
	return &CleanupProcessor{
		db:           db,
		queueService: queueService,
		logger:       logger,
		stopChan:     make(chan struct{}),
	}
}

// Name returns the processor name
func (p *CleanupProcessor) Name() string {
	return "CleanupProcessor"
}

// Run starts the processor loop
func (p *CleanupProcessor) Run(ctx context.Context) error {
	p.logger.Info("CleanupProcessor started (runs daily at 2 AM)")

	// Calculate time until next 2 AM
	now := time.Now()
	next2AM := time.Date(now.Year(), now.Month(), now.Day(), 2, 0, 0, 0, now.Location())
	if now.After(next2AM) {
		next2AM = next2AM.Add(24 * time.Hour)
	}

	// Wait until 2 AM
	timer := time.NewTimer(time.Until(next2AM))
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			p.logger.Info("CleanupProcessor stopping (context cancelled)")
			return nil
		case <-p.stopChan:
			p.logger.Info("CleanupProcessor stopped")
			return nil
		case <-timer.C:
			// Run cleanup tasks
			if err := p.runCleanup(ctx); err != nil {
				p.logger.Error(fmt.Sprintf("Cleanup failed: %v", err))
			}

			// Schedule next run (24 hours from now)
			timer.Reset(24 * time.Hour)
		}
	}
}

// Stop gracefully stops the processor
func (p *CleanupProcessor) Stop(ctx context.Context) error {
	p.logger.Info("Stopping CleanupProcessor...")
	close(p.stopChan)
	return nil
}

// runCleanup performs all cleanup tasks
func (p *CleanupProcessor) runCleanup(ctx context.Context) error {
	p.logger.Info("🧹 Starting daily cleanup tasks...")

	startTime := time.Now()

	// Task 1: Delete old draft posts (30+ days)
	if err := p.cleanupOldDrafts(ctx); err != nil {
		p.logger.Error(fmt.Sprintf("Failed to cleanup old drafts: %v", err))
	}

	// Task 2: Delete expired refresh tokens
	if err := p.cleanupExpiredTokens(ctx); err != nil {
		p.logger.Error(fmt.Sprintf("Failed to cleanup expired tokens: %v", err))
	}

	// Task 3: Archive old analytics (1 year+)
	if err := p.archiveOldAnalytics(ctx); err != nil {
		p.logger.Error(fmt.Sprintf("Failed to archive old analytics: %v", err))
	}

	// Task 4: Clean up dead letter queue
	if err := p.cleanupDLQ(ctx); err != nil {
		p.logger.Error(fmt.Sprintf("Failed to cleanup DLQ: %v", err))
	}

	// Task 5: Vacuum database (optional, for PostgreSQL)
	if err := p.vacuumDatabase(ctx); err != nil {
		p.logger.Error(fmt.Sprintf("Failed to vacuum database: %v", err))
	}

	duration := time.Since(startTime)
	p.logger.Info(fmt.Sprintf("✅ Cleanup completed in %v", duration))

	return nil
}

// cleanupOldDrafts deletes draft posts older than 30 days
func (p *CleanupProcessor) cleanupOldDrafts(ctx context.Context) error {
	p.logger.Info("Cleaning up old draft posts (30+ days)...")

	cutoffDate := time.Now().AddDate(0, 0, -30)

	query := `
		UPDATE scheduled_posts
		SET deleted_at = NOW()
		WHERE status = 'draft'
		  AND created_at < $1
		  AND deleted_at IS NULL
	`

	result, err := p.db.ExecContext(ctx, query, cutoffDate)
	if err != nil {
		return fmt.Errorf("failed to delete old drafts: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	p.logger.Info(fmt.Sprintf("Deleted %d old draft posts", rowsAffected))

	return nil
}

// cleanupExpiredTokens deletes expired refresh tokens
func (p *CleanupProcessor) cleanupExpiredTokens(ctx context.Context) error {
	p.logger.Info("Cleaning up expired refresh tokens...")

	query := `
		DELETE FROM refresh_tokens
		WHERE expires_at < NOW()
	`

	result, err := p.db.ExecContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to delete expired tokens: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	p.logger.Info(fmt.Sprintf("Deleted %d expired refresh tokens", rowsAffected))

	return nil
}

// archiveOldAnalytics archives analytics events older than 1 year
func (p *CleanupProcessor) archiveOldAnalytics(ctx context.Context) error {
	p.logger.Info("Archiving old analytics (1 year+)...")

	cutoffDate := time.Now().AddDate(-1, 0, 0)

	// In production, you'd move to an archive table or S3
	// For now, we'll just delete
	query := `
		DELETE FROM analytics_events
		WHERE timestamp < $1
	`

	result, err := p.db.ExecContext(ctx, query, cutoffDate)
	if err != nil {
		return fmt.Errorf("failed to archive analytics: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	p.logger.Info(fmt.Sprintf("Archived %d analytics events", rowsAffected))

	return nil
}

// cleanupDLQ processes and clears the dead letter queue
func (p *CleanupProcessor) cleanupDLQ(ctx context.Context) error {
	p.logger.Info("Cleaning up dead letter queue...")

	jobTypes := []string{"publish_post", "fetch_analytics"}

	totalCleaned := 0
	for _, jobType := range jobTypes {
		dlqLength, err := p.queueService.GetDLQLength(ctx, jobType)
		if err != nil {
			p.logger.Warn(fmt.Sprintf("Failed to get DLQ length for %s: %v", jobType, err))
			continue
		}

		if dlqLength > 0 {
			// In production, you'd log these to a monitoring system
			p.logger.Warn(fmt.Sprintf("Found %d permanently failed jobs in DLQ: %s", dlqLength, jobType))

			// Optionally purge old DLQ items (older than 7 days)
			// For now, we'll just log
			totalCleaned += int(dlqLength)
		}
	}

	p.logger.Info(fmt.Sprintf("DLQ cleanup complete: %d failed jobs logged", totalCleaned))
	return nil
}

// vacuumDatabase runs VACUUM ANALYZE on PostgreSQL
func (p *CleanupProcessor) vacuumDatabase(ctx context.Context) error {
	p.logger.Info("Running database vacuum...")

	// VACUUM cannot run inside a transaction
	_, err := p.db.ExecContext(ctx, "VACUUM ANALYZE")
	if err != nil {
		return fmt.Errorf("failed to vacuum database: %w", err)
	}

	p.logger.Info("Database vacuum completed")
	return nil
}

// Additional utility methods

// cleanupOldJobRuns deletes job run records older than 30 days
func (p *CleanupProcessor) cleanupOldJobRuns(ctx context.Context) error {
	p.logger.Info("Cleaning up old job runs (30+ days)...")

	cutoffDate := time.Now().AddDate(0, 0, -30)

	query := `
		DELETE FROM job_runs
		WHERE created_at < $1
		  AND status IN ('completed', 'failed')
	`

	result, err := p.db.ExecContext(ctx, query, cutoffDate)
	if err != nil {
		return fmt.Errorf("failed to cleanup old job runs: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	p.logger.Info(fmt.Sprintf("Deleted %d old job run records", rowsAffected))

	return nil
}

// cleanupUnverifiedUsers deletes unverified users after 7 days
func (p *CleanupProcessor) cleanupUnverifiedUsers(ctx context.Context) error {
	p.logger.Info("Cleaning up unverified users (7+ days)...")

	cutoffDate := time.Now().AddDate(0, 0, -7)

	query := `
		UPDATE users
		SET deleted_at = NOW()
		WHERE email_verified = false
		  AND created_at < $1
		  AND deleted_at IS NULL
	`

	result, err := p.db.ExecContext(ctx, query, cutoffDate)
	if err != nil {
		return fmt.Errorf("failed to cleanup unverified users: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	p.logger.Info(fmt.Sprintf("Deleted %d unverified users", rowsAffected))

	return nil
}
