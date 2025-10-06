-- path: backend/sql/posts.sql

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
  AND p.published_at BETWEEN sqlc.arg('start_date') AND sqlc.arg('end_date')
ORDER BY p.published_at DESC
LIMIT $2 OFFSET $3;

-- name: UpdatePostAnalytics :exec
UPDATE posts
SET 
    impressions = $2,
    engagements = $3,
    clicks = $4,
    likes = $5,
    shares = $6,
    comments = $7,
    last_analytics_fetch_at = NOW(),
    updated_at = NOW()
WHERE id = $1;

-- name: IncrementPostMetric :exec
UPDATE posts
SET 
    impressions = CASE WHEN sqlc.arg('metric') = 'impressions' THEN impressions + sqlc.arg('value') ELSE impressions END,
    engagements = CASE WHEN sqlc.arg('metric') = 'engagements' THEN engagements + sqlc.arg('value') ELSE engagements END,
    clicks = CASE WHEN sqlc.arg('metric') = 'clicks' THEN clicks + sqlc.arg('value') ELSE clicks END,
    likes = CASE WHEN sqlc.arg('metric') = 'likes' THEN likes + sqlc.arg('value') ELSE likes END,
    shares = CASE WHEN sqlc.arg('metric') = 'shares' THEN shares + sqlc.arg('value') ELSE shares END,
    comments = CASE WHEN sqlc.arg('metric') = 'comments' THEN comments + sqlc.arg('value') ELSE comments END,
    updated_at = NOW()
WHERE id = sqlc.arg('post_id');