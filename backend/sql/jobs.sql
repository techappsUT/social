-- path: backend/sql/jobs.sql

-- name: CreateJobRun :one
INSERT INTO job_runs (
    job_name,
    job_type,
    status,
    context,
    max_attempts
) VALUES (
    $1, $2, $3, $4, $5
)
RETURNING *;

-- name: GetJobRun :one
SELECT * FROM job_runs WHERE id = $1;

-- name: StartJobRun :exec
UPDATE job_runs
SET 
    status = 'running',
    started_at = NOW(),
    worker_id = $2,
    updated_at = NOW()
WHERE id = $1;

-- name: CompleteJobRun :exec
UPDATE job_runs
SET 
    status = 'completed',
    completed_at = NOW(),
    duration_ms = EXTRACT(EPOCH FROM (NOW() - started_at)) * 1000,
    result = sqlc.narg('result'),
    updated_at = NOW()
WHERE id = $1;

-- name: FailJobRun :exec
UPDATE job_runs
SET 
    status = 'failed',
    completed_at = NOW(),
    duration_ms = EXTRACT(EPOCH FROM (NOW() - started_at)) * 1000,
    error_message = $2,
    stack_trace = sqlc.narg('stack_trace'),
    attempt_number = attempt_number + 1,
    next_retry_at = CASE 
        WHEN attempt_number < max_attempts 
        THEN NOW() + (INTERVAL '1 minute' * POWER(2, attempt_number))
        ELSE NULL
    END,
    updated_at = NOW()
WHERE id = $1;

-- name: ListPendingRetries :many
SELECT * FROM job_runs
WHERE status = 'failed'
  AND next_retry_at IS NOT NULL
  AND next_retry_at <= NOW()
  AND attempt_number < max_attempts
ORDER BY next_retry_at ASC
LIMIT $1;

-- name: ListRecentJobRuns :many
SELECT * FROM job_runs
WHERE job_name = $1
ORDER BY created_at DESC
LIMIT $2;