-- path: backend/sql/post_queue.sql

-- name: EnqueuePost :one
INSERT INTO post_queue (
    scheduled_post_id,
    status,
    priority,
    scheduled_for,
    max_attempts
) VALUES (
    $1, $2, $3, $4, $5
)
RETURNING *;

-- name: GetQueueItem :one
SELECT * FROM post_queue
WHERE id = $1;

-- name: LockNextQueueItems :many
UPDATE post_queue SET 
    status = 'processing',
    worker_id = $1,
    locked_at = NOW(),
    lock_expires_at = NOW() + INTERVAL '5 minutes',
    updated_at = NOW()
WHERE id IN (
    SELECT pq.id 
    FROM post_queue pq
    WHERE pq.status = 'pending'
      AND pq.scheduled_for <= $2
      AND (pq.lock_expires_at IS NULL OR pq.lock_expires_at < NOW())
    ORDER BY pq.priority ASC, pq.scheduled_for ASC
    LIMIT $3
    FOR UPDATE SKIP LOCKED
)
RETURNING *;

-- name: CompleteQueueItem :exec
UPDATE post_queue SET 
    status = 'completed',
    completed_at = NOW(),
    updated_at = NOW()
WHERE id = $1;

-- name: FailQueueItem :exec
UPDATE post_queue SET 
    status = CASE 
        WHEN attempt_count + 1 >= max_attempts THEN 'failed'::queue_status
        ELSE 'retrying'::queue_status
    END,
    attempt_count = attempt_count + 1,
    last_error = $2,
    updated_at = NOW()
WHERE id = $1;

-- name: ResetStaleLocks :exec
UPDATE post_queue SET 
    status = 'pending',
    worker_id = NULL,
    locked_at = NULL,
    lock_expires_at = NULL,
    updated_at = NOW()
WHERE status = 'processing'
  AND lock_expires_at < NOW();