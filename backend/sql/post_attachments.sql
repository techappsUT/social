-- path: backend/sql/post_attachments.sql

-- name: CreatePostAttachment :one
INSERT INTO post_attachments (
    scheduled_post_id,
    type,
    url,
    thumbnail_url,
    file_size,
    mime_type,
    width,
    height,
    duration,
    alt_text,
    display_order
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
)
RETURNING *;

-- name: ListPostAttachments :many
SELECT * FROM post_attachments
WHERE scheduled_post_id = $1
ORDER BY display_order ASC, created_at ASC;

-- name: DeletePostAttachment :exec
DELETE FROM post_attachments
WHERE id = $1;

-- name: DeletePostAttachmentsByScheduledPost :exec
DELETE FROM post_attachments
WHERE scheduled_post_id = $1;