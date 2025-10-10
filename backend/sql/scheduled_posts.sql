-- path: backend/sql/scheduled_posts.sql
-- ✅ KEEP - Verify this file exists with these queries

-- name: CreateScheduledPost :one
INSERT INTO scheduled_posts (
    team_id,
    created_by,
    social_account_id,
    content,
    content_html,
    shortened_links,
    status,
    scheduled_at,
    platform_specific_options
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
)
RETURNING *;

-- name: GetScheduledPostByID :one
SELECT * FROM scheduled_posts
WHERE id = $1 AND deleted_at IS NULL;

-- name: ListScheduledPostsByTeam :many
SELECT * FROM scheduled_posts
WHERE team_id = $1 
  AND deleted_at IS NULL
ORDER BY scheduled_at DESC
LIMIT $2 OFFSET $3;

-- name: GetDuePosts :many
SELECT sp.*, sa.platform
FROM scheduled_posts sp
INNER JOIN social_accounts sa ON sp.social_account_id = sa.id
WHERE sp.status IN ('scheduled', 'queued')
  AND sp.scheduled_at <= $1
  AND sp.deleted_at IS NULL
  AND sa.status = 'active'
  AND sa.deleted_at IS NULL
ORDER BY sp.scheduled_at ASC
LIMIT $2;

-- name: UpdateScheduledPostStatus :exec
UPDATE scheduled_posts
SET 
    status = $2,
    error_message = COALESCE(sqlc.narg('error_message'), error_message),
    retry_count = COALESCE(sqlc.narg('retry_count'), retry_count),
    published_at = COALESCE(sqlc.narg('published_at'), published_at),
    updated_at = NOW()
WHERE id = $1;

-- name: UpdateScheduledPost :one
UPDATE scheduled_posts
SET 
    content = COALESCE(sqlc.narg('content'), content),
    content_html = COALESCE(sqlc.narg('content_html'), content_html),
    scheduled_at = COALESCE(sqlc.narg('scheduled_at'), scheduled_at),
    platform_specific_options = COALESCE(sqlc.narg('platform_specific_options'), platform_specific_options),
    updated_at = NOW()
WHERE id = sqlc.arg('id') AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeleteScheduledPost :exec
UPDATE scheduled_posts
SET 
    deleted_at = NOW(),
    updated_at = NOW()
WHERE id = $1;

-- name: CountScheduledPostsByTeam :one
SELECT COUNT(*)
FROM scheduled_posts
WHERE team_id = $1 
  AND deleted_at IS NULL;

-- name: IncrementRetryCount :exec
UPDATE scheduled_posts
SET 
    retry_count = retry_count + 1,
    error_message = $2,
    updated_at = NOW()
WHERE id = $1;