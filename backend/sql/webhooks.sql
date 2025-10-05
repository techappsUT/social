-- path: backend/sql/webhooks.sql

-- name: CreateWebhookLog :one
INSERT INTO webhooks_log (
    source,
    event_type,
    payload,
    headers,
    idempotency_key
) VALUES (
    $1, $2, $3, $4, $5
)
RETURNING *;

-- name: GetWebhookLog :one
SELECT * FROM webhooks_log WHERE id = $1;

-- name: MarkWebhookProcessed :exec
UPDATE webhooks_log
SET 
    processed = true,
    processed_at = NOW(),
    response_status = sqlc.narg('response_status'),
    response_body = sqlc.narg('response_body')
WHERE id = $1;

-- name: MarkWebhookFailed :exec
UPDATE webhooks_log
SET 
    processed = false,
    error_message = $2
WHERE id = $1;

-- name: ListUnprocessedWebhooks :many
SELECT * FROM webhooks_log
WHERE processed = false
ORDER BY created_at ASC
LIMIT $1;