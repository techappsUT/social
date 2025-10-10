-- path: backend/sql/jobs.sql
-- 🔄 REFACTORED - Match actual job_runs schema

-- name: CreateJobRun :one
INSERT INTO job_runs (
    job_name,
    status,
    payload,
    started_at
) VALUES (
    $1, $2, $3, $4
)
RETURNING *;

-- name: GetJobRunByID :one
SELECT * FROM job_runs WHERE id = $1;

-- name: UpdateJobRunStatus :exec
UPDATE job_runs
SET 
    status = $2,
    updated_at = NOW()
WHERE id = $1;

-- name: CompleteJobRun :exec
UPDATE job_runs
SET 
    status = 'completed',
    result = $2,
    completed_at = NOW(),
    updated_at = NOW()
WHERE id = $1;

-- name: FailJobRun :exec
UPDATE job_runs
SET 
    status = 'failed',
    error = $2,
    completed_at = NOW(),
    updated_at = NOW()
WHERE id = $1;

-- name: ListRecentJobRuns :many
SELECT *
FROM job_runs
WHERE job_name = $1
ORDER BY created_at DESC
LIMIT $2;

-- name: ListFailedJobRuns :many
SELECT *
FROM job_runs
WHERE status = 'failed'
  AND created_at >= $1
ORDER BY created_at DESC
LIMIT $2;