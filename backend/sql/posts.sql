-- path: backend/sql/posts.sql
-- 🔄 REFACTORED - Schema has impressions in analytics_events, not posts table

-- name: CreatePost :one
INSERT INTO posts (
    scheduled_post_id,
    team_id,
    social_account_id,
    platform_post_id,
    platform_post_url,
    content,
    published_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
)
RETURNING *;

-- name: GetPostByID :one
SELECT * FROM posts WHERE id = $1;

-- name: GetPostByScheduledPostID :one
SELECT * FROM posts WHERE scheduled_post_id = $1;

-- name: GetPostByPlatformID :one
SELECT * FROM posts
WHERE social_account_id = $1 AND platform_post_id = $2;

-- name: ListPostsByTeam :many
SELECT 
    p.*,
    sa.platform,
    sa.username
FROM posts p
INNER JOIN social_accounts sa ON p.social_account_id = sa.id
WHERE p.team_id = $1
  AND p.published_at BETWEEN $2 AND $3
ORDER BY p.published_at DESC
LIMIT $4 OFFSET $5;

-- name: ListRecentPostsByTeam :many
SELECT 
    p.*,
    sa.platform,
    sa.username
FROM posts p
INNER JOIN social_accounts sa ON p.social_account_id = sa.id
WHERE p.team_id = $1
ORDER BY p.published_at DESC
LIMIT $2;

-- name: CountPostsByTeam :one
SELECT COUNT(*)
FROM posts
WHERE team_id = $1;

-- name: CountPostsByTeamAndDateRange :one
SELECT COUNT(*)
FROM posts
WHERE team_id = $1
  AND published_at BETWEEN $2 AND $3;