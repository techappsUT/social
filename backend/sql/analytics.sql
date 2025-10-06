-- path: backend/sql/analytics.sql

-- name: InsertAnalyticsEvent :one
INSERT INTO analytics_events (
    post_id,
    team_id,
    event_type,
    event_value,
    platform,
    country,
    device_type,
    referrer,
    event_metadata,
    event_timestamp
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
)
RETURNING *;

-- name: GetAnalyticsByPost :many
SELECT 
    event_type,
    SUM(event_value) as total_value,
    COUNT(*) as event_count
FROM analytics_events
WHERE post_id = $1
GROUP BY event_type;

-- name: GetTeamAnalyticsSummary :one
SELECT 
    COUNT(DISTINCT p.id) as total_posts,
    COALESCE(SUM(p.impressions), 0) as total_impressions,
    COALESCE(SUM(p.engagements), 0) as total_engagements,
    COALESCE(SUM(p.clicks), 0) as total_clicks,
    COALESCE(SUM(p.likes), 0) as total_likes,
    COALESCE(SUM(p.shares), 0) as total_shares,
    COALESCE(SUM(p.comments), 0) as total_comments
FROM posts p
WHERE p.team_id = $1
  AND p.published_at BETWEEN $2 AND $3;

-- name: GetAnalyticsByPlatform :many
SELECT 
    sa.platform,
    COUNT(p.id) as post_count,
    COALESCE(SUM(p.impressions), 0) as total_impressions,
    COALESCE(SUM(p.engagements), 0) as total_engagements,
    COALESCE(SUM(p.clicks), 0) as total_clicks
FROM posts p
INNER JOIN social_accounts sa ON p.social_account_id = sa.id
WHERE p.team_id = $1
  AND p.published_at BETWEEN $2 AND $3
GROUP BY sa.platform
ORDER BY total_engagements DESC;

-- name: GetTopPerformingPosts :many
SELECT 
    p.*,
    sa.platform,
    sa.username
FROM posts p
INNER JOIN social_accounts sa ON p.social_account_id = sa.id
WHERE p.team_id = $1
  AND p.published_at BETWEEN $2 AND $3
ORDER BY 
    CASE 
        WHEN sqlc.arg('sort_by') = 'impressions' THEN p.impressions
        WHEN sqlc.arg('sort_by') = 'engagements' THEN p.engagements
        WHEN sqlc.arg('sort_by') = 'clicks' THEN p.clicks
        ELSE p.engagements
    END DESC
LIMIT $4;

-- name: GetAnalyticsTimeSeries :many
SELECT 
    DATE(p.published_at) as date,
    COUNT(p.id) as posts,
    COALESCE(SUM(p.impressions), 0) as impressions,
    COALESCE(SUM(p.engagements), 0) as engagements,
    COALESCE(SUM(p.clicks), 0) as clicks
FROM posts p
WHERE p.team_id = $1
  AND p.published_at BETWEEN $2 AND $3
GROUP BY DATE(p.published_at)
ORDER BY date ASC;