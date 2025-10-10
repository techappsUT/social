-- path: backend/sql/post_queue.sql
-- 🔄 REFACTORED - Use max_attempts and error (not max_retries, error_message)

-- name: EnqueuePost :one
INSERT INTO post_queue (
    scheduled_post_id,
    priority,
    scheduled_for,
    max_attempts
) VALUES (
    $1, $2, $3, $4
)
RETURNING *;

-- name: GetNextQueuedPosts :many
SELECT * FROM post_queue
WHERE status = 'pending'
  AND scheduled_for <= NOW()
ORDER BY priority DESC, scheduled_for ASC
LIMIT $1
FOR UPDATE SKIP LOCKED;

-- name: LockQueueItem :exec
UPDATE post_queue
SET 
    status = 'processing',
    started_at = NOW(),
    attempts = attempts + 1,
    updated_at = NOW()
WHERE id = $1;

-- name: CompleteQueueItem :exec
UPDATE post_queue
SET 
    status = 'completed',
    completed_at = NOW(),
    updated_at = NOW()
WHERE id = $1;

-- name: FailQueueItem :exec
UPDATE post_queue
SET 
    status = CASE 
        WHEN attempts >= max_attempts THEN 'failed'::queue_status
        ELSE 'pending'::queue_status
    END,
    error = $2,
    updated_at = NOW()
WHERE id = $1;

-- name: GetQueueItemByID :one
SELECT * FROM post_queue WHERE id = $1;

-- name: ListQueuedPostsByStatus :many
SELECT * FROM post_queue
WHERE status = $1
ORDER BY scheduled_for DESC
LIMIT $2 OFFSET $3;

-- name: CountQueuedPostsByStatus :one
SELECT COUNT(*) FROM post_queue
WHERE status = $1;

-- name: RetryFailedQueueItem :exec
UPDATE post_queue
SET 
    status = 'pending',
    error = NULL,
    attempts = 0,
    updated_at = NOW()
WHERE id = $1;

-- name: ListPendingQueueItems :many
SELECT 
    pq.*,
    sp.content,
    sa.platform,
    sa.username
FROM post_queue pq
INNER JOIN scheduled_posts sp ON pq.scheduled_post_id = sp.id
INNER JOIN social_accounts sa ON sp.social_account_id = sa.id
WHERE pq.status = 'pending'
  AND pq.scheduled_for <= $1
ORDER BY pq.priority DESC, pq.scheduled_for ASC
LIMIT $2;