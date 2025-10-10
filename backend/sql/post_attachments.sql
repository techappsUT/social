-- path: backend/sql/post_attachments.sql
-- 🔄 REFACTORED - Match actual schema (url, upload_order, type)

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
    upload_order
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
)
RETURNING *;

-- name: GetPostAttachmentByID :one
SELECT * FROM post_attachments WHERE id = $1;

-- name: ListPostAttachmentsByScheduledPost :many
SELECT * FROM post_attachments
WHERE scheduled_post_id = $1
ORDER BY upload_order ASC, created_at ASC;

-- name: UpdatePostAttachment :one
UPDATE post_attachments
SET 
    url = COALESCE(sqlc.narg('url'), url),
    thumbnail_url = COALESCE(sqlc.narg('thumbnail_url'), thumbnail_url),
    alt_text = COALESCE(sqlc.narg('alt_text'), alt_text),
    width = COALESCE(sqlc.narg('width'), width),
    height = COALESCE(sqlc.narg('height'), height),
    upload_order = COALESCE(sqlc.narg('upload_order'), upload_order)
WHERE id = sqlc.arg('id')
RETURNING *;

-- name: DeletePostAttachment :exec
DELETE FROM post_attachments WHERE id = $1;

-- name: DeletePostAttachmentsByScheduledPost :exec
DELETE FROM post_attachments WHERE scheduled_post_id = $1;

-- name: CountAttachmentsByPost :one
SELECT COUNT(*) FROM post_attachments
WHERE scheduled_post_id = $1;

-- name: GetAttachmentsByType :many
SELECT * FROM post_attachments
WHERE scheduled_post_id = $1 AND type = $2
ORDER BY upload_order ASC;