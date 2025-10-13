// ============================================================================
// FILE: backend/internal/infrastructure/services/worker_queue.go
// PURPOSE: Redis-based job queue with retry and DLQ
// ============================================================================

package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/techappsUT/social-queue/internal/application/common"
)

const (
	MaxRetries          = 3
	QueueKeyPrefix      = "queue:"
	ProcessingKeyPrefix = "processing:"
	DLQKeyPrefix        = "dlq:"
	JobDataKeyPrefix    = "job:data:"
)

// Job represents a background job
type Job struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"`
	Payload    map[string]interface{} `json:"payload"`
	CreatedAt  time.Time              `json:"created_at"`
	RetryCount int                    `json:"retry_count"`
	LastError  string                 `json:"last_error,omitempty"`
}

// WorkerQueueService implements job queue using Redis
type WorkerQueueService struct {
	client *redis.Client
	logger common.Logger
}

// NewWorkerQueueService creates a new worker queue service
func NewWorkerQueueService(client *redis.Client, logger common.Logger) *WorkerQueueService {
	return &WorkerQueueService{
		client: client,
		logger: logger,
	}
}

// Enqueue adds a job to the queue
func (w *WorkerQueueService) Enqueue(ctx context.Context, jobType string, payload map[string]interface{}) (string, error) {
	job := &Job{
		ID:         uuid.New().String(),
		Type:       jobType,
		Payload:    payload,
		CreatedAt:  time.Now().UTC(),
		RetryCount: 0,
	}

	// Serialize job
	jobData, err := json.Marshal(job)
	if err != nil {
		return "", fmt.Errorf("failed to marshal job: %w", err)
	}

	// Store job data with 24 hour expiry
	jobDataKey := fmt.Sprintf("%s%s", JobDataKeyPrefix, job.ID)
	if err := w.client.Set(ctx, jobDataKey, jobData, 24*time.Hour).Err(); err != nil {
		return "", fmt.Errorf("failed to store job data: %w", err)
	}

	// Add job ID to queue
	queueKey := fmt.Sprintf("%s%s", QueueKeyPrefix, jobType)
	if err := w.client.RPush(ctx, queueKey, job.ID).Err(); err != nil {
		return "", fmt.Errorf("failed to enqueue job: %w", err)
	}

	w.logger.Info(fmt.Sprintf("Enqueued job: %s (type: %s)", job.ID, jobType))
	return job.ID, nil
}

// Dequeue retrieves and locks a job from the queue (blocks until available)
func (w *WorkerQueueService) Dequeue(ctx context.Context, jobType string, timeout time.Duration) (*Job, error) {
	queueKey := fmt.Sprintf("%s%s", QueueKeyPrefix, jobType)
	processingKey := fmt.Sprintf("%s%s", ProcessingKeyPrefix, jobType)

	// Use BRPOPLPUSH for atomic dequeue with timeout
	jobID, err := w.client.BRPopLPush(ctx, queueKey, processingKey, timeout).Result()
	if err == redis.Nil {
		return nil, nil // No jobs available
	}
	if err != nil {
		return nil, fmt.Errorf("failed to dequeue job: %w", err)
	}

	// Get job data
	jobDataKey := fmt.Sprintf("%s%s", JobDataKeyPrefix, jobID)
	jobData, err := w.client.Get(ctx, jobDataKey).Result()
	if err == redis.Nil {
		// Job data expired, remove from processing
		w.client.LRem(ctx, processingKey, 1, jobID)
		return nil, fmt.Errorf("job data not found: %s", jobID)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get job data: %w", err)
	}

	// Deserialize job
	var job Job
	if err := json.Unmarshal([]byte(jobData), &job); err != nil {
		return nil, fmt.Errorf("failed to unmarshal job: %w", err)
	}

	w.logger.Info(fmt.Sprintf("Dequeued job: %s (type: %s)", job.ID, jobType))
	return &job, nil
}

// MarkComplete marks a job as successfully completed
func (w *WorkerQueueService) MarkComplete(ctx context.Context, jobType string, jobID string) error {
	processingKey := fmt.Sprintf("%s%s", ProcessingKeyPrefix, jobType)

	// Remove from processing list
	if err := w.client.LRem(ctx, processingKey, 1, jobID).Err(); err != nil {
		return fmt.Errorf("failed to remove from processing: %w", err)
	}

	// Delete job data
	jobDataKey := fmt.Sprintf("%s%s", JobDataKeyPrefix, jobID)
	if err := w.client.Del(ctx, jobDataKey).Err(); err != nil {
		w.logger.Warn(fmt.Sprintf("Failed to delete job data: %s", jobID))
	}

	w.logger.Info(fmt.Sprintf("Completed job: %s", jobID))
	return nil
}

// MarkFailed marks a job as failed and handles retry logic
func (w *WorkerQueueService) MarkFailed(ctx context.Context, jobType string, jobID string, errorMsg string) error {
	processingKey := fmt.Sprintf("%s%s", ProcessingKeyPrefix, jobType)
	jobDataKey := fmt.Sprintf("%s%s", JobDataKeyPrefix, jobID)

	// Get current job data
	jobData, err := w.client.Get(ctx, jobDataKey).Result()
	if err != nil {
		return fmt.Errorf("failed to get job data: %w", err)
	}

	var job Job
	if err := json.Unmarshal([]byte(jobData), &job); err != nil {
		return fmt.Errorf("failed to unmarshal job: %w", err)
	}

	// Update job with error info
	job.RetryCount++
	job.LastError = errorMsg

	// Check if we should retry
	if job.RetryCount <= MaxRetries {
		// Calculate exponential backoff: 1min, 10min, 1hour
		backoffDuration := time.Duration(1<<uint(job.RetryCount-1)) * 10 * time.Minute

		w.logger.Warn(fmt.Sprintf("Job %s failed (retry %d/%d), retrying in %v: %s",
			jobID, job.RetryCount, MaxRetries, backoffDuration, errorMsg))

		// Update job data
		updatedData, _ := json.Marshal(job)
		w.client.Set(ctx, jobDataKey, updatedData, 24*time.Hour)

		// Schedule retry (simplified: just re-add to queue immediately)
		// In production, use Redis sorted set with score = retry_timestamp
		queueKey := fmt.Sprintf("%s%s", QueueKeyPrefix, jobType)
		w.client.RPush(ctx, queueKey, jobID)
	} else {
		// Max retries exceeded, move to DLQ
		w.logger.Error(fmt.Sprintf("Job %s permanently failed after %d retries: %s",
			jobID, MaxRetries, errorMsg))

		dlqKey := fmt.Sprintf("%s%s", DLQKeyPrefix, jobType)

		// Add to DLQ with job data
		updatedData, _ := json.Marshal(job)
		w.client.RPush(ctx, dlqKey, string(updatedData))
	}

	// Remove from processing list
	w.client.LRem(ctx, processingKey, 1, jobID)
	return nil
}

// GetQueueLength returns the number of jobs in a queue
func (w *WorkerQueueService) GetQueueLength(ctx context.Context, jobType string) (int64, error) {
	queueKey := fmt.Sprintf("%s%s", QueueKeyPrefix, jobType)
	length, err := w.client.LLen(ctx, queueKey).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get queue length: %w", err)
	}
	return length, nil
}

// GetProcessingLength returns the number of jobs being processed
func (w *WorkerQueueService) GetProcessingLength(ctx context.Context, jobType string) (int64, error) {
	processingKey := fmt.Sprintf("%s%s", ProcessingKeyPrefix, jobType)
	length, err := w.client.LLen(ctx, processingKey).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get processing length: %w", err)
	}
	return length, nil
}

// GetDLQLength returns the number of permanently failed jobs
func (w *WorkerQueueService) GetDLQLength(ctx context.Context, jobType string) (int64, error) {
	dlqKey := fmt.Sprintf("%s%s", DLQKeyPrefix, jobType)
	length, err := w.client.LLen(ctx, dlqKey).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get DLQ length: %w", err)
	}
	return length, nil
}

// PurgeQueue removes all jobs from a queue (use with caution!)
func (w *WorkerQueueService) PurgeQueue(ctx context.Context, jobType string) error {
	queueKey := fmt.Sprintf("%s%s", QueueKeyPrefix, jobType)
	if err := w.client.Del(ctx, queueKey).Err(); err != nil {
		return fmt.Errorf("failed to purge queue: %w", err)
	}
	w.logger.Warn(fmt.Sprintf("Purged queue: %s", jobType))
	return nil
}
