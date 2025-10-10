-- path: backend/sql/webhooks.sql
-- 🔄 REFACTORED - Match actual schema (webhooks_log table)

-- name: CreateWebhookLog :one
INSERT INTO webhooks_log (
    source,
    event_type,
    payload,
    processed
) VALUES (
    $1, $2, $3, $4
)
RETURNING *;

-- name: GetWebhookLogByID :one
SELECT * FROM webhooks_log WHERE id = $1;

-- name: ListWebhookLogs :many
SELECT * FROM webhooks_log
WHERE source = $1
ORDER BY received_at DESC
LIMIT $2 OFFSET $3;

-- name: ListUnprocessedWebhooks :many
SELECT * FROM webhooks_log
WHERE processed = FALSE
ORDER BY received_at ASC
LIMIT $1;

-- name: MarkWebhookProcessed :exec
UPDATE webhooks_log
SET processed = TRUE
WHERE id = $1;

-- name: CountWebhooksBySource :one
SELECT COUNT(*)
FROM webhooks_log
WHERE source = $1;