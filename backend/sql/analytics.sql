-- path: backend/sql/analytics.sql
-- 🔄 REFACTORED - Match actual schema (recorded_at, no team_id, no event_timestamp)

-- name: CreateAnalyticsEvent :one
INSERT INTO analytics_events (
    post_id,
    event_type,
    event_value,
    event_metadata
) VALUES (
    $1, $2, $3, $4
)
RETURNING *;

-- name: GetAnalyticsEventsByPost :many
SELECT * FROM analytics_events
WHERE post_id = $1
ORDER BY recorded_at DESC;

-- name: GetAnalyticsEventsByDateRange :many
SELECT ae.*
FROM analytics_events ae
INNER JOIN posts p ON ae.post_id = p.id
WHERE p.team_id = $1
  AND ae.recorded_at BETWEEN $2 AND $3
ORDER BY ae.recorded_at DESC
LIMIT $4 OFFSET $5;

-- name: GetAnalyticsSummaryByTeam :one
SELECT 
    COUNT(DISTINCT ae.post_id) as total_posts,
    COUNT(*) as total_events,
    COUNT(*) FILTER (WHERE ae.event_type = 'impression') as total_impressions,
    COUNT(*) FILTER (WHERE ae.event_type IN ('like', 'share', 'comment')) as total_engagements,
    COUNT(*) FILTER (WHERE ae.event_type = 'click') as total_clicks
FROM analytics_events ae
INNER JOIN posts p ON ae.post_id = p.id
WHERE p.team_id = $1
  AND ae.recorded_at BETWEEN $2 AND $3;

-- name: GetEventCountByType :many
SELECT 
    event_type,
    COUNT(*) as event_count,
    SUM(event_value) as total_value
FROM analytics_events
WHERE post_id = $1
GROUP BY event_type
ORDER BY total_value DESC;