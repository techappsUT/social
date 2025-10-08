-- path: backend/internal/social/queries/social_tokens.sql

-- name: CreateSocialToken :one
INSERT INTO social_tokens (
    user_id,
    platform_type,
    platform_user_id,
    platform_username,
    access_token,
    refresh_token,
    expires_at,
    scope,
    extra
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
)
RETURNING *;

-- name: GetSocialToken :one
SELECT * FROM social_tokens
WHERE id = $1 AND is_valid = true
LIMIT 1;

-- name: GetSocialTokensByUser :many
SELECT * FROM social_tokens
WHERE user_id = $1 AND is_valid = true
ORDER BY created_at DESC;

-- name: GetSocialTokenByPlatform :one
SELECT * FROM social_tokens
WHERE user_id = $1 
  AND platform_type = $2 
  AND platform_user_id = $3
  AND is_valid = true
LIMIT 1;

-- name: UpdateSocialToken :one
UPDATE social_tokens
SET access_token = $2,
    refresh_token = $3,
    expires_at = $4,
    last_validated = NOW(),
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: InvalidateSocialToken :exec
UPDATE social_tokens
SET is_valid = false,
    updated_at = NOW()
WHERE id = $1;

-- name: DeleteSocialToken :exec
DELETE FROM social_tokens
WHERE id = $1;

-- name: GetExpiringSocialTokens :many
SELECT * FROM social_tokens
WHERE expires_at <= $1 
  AND is_valid = true
  AND refresh_token IS NOT NULL
ORDER BY expires_at ASC
LIMIT $2;