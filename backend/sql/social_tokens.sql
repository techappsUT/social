-- path: backend/sql/social_tokens.sql
-- 🔄 REFACTORED - Fixed syntax and removed duplicates

-- name: CreateSocialToken :one
INSERT INTO social_tokens (
    social_account_id,
    access_token,
    refresh_token,
    token_type,
    expires_at,
    scope
) VALUES (
    $1, $2, $3, $4, $5, $6
)
RETURNING *;

-- name: GetSocialTokenByAccountID :one
SELECT * FROM social_tokens
WHERE social_account_id = $1;

-- name: UpdateSocialToken :exec
UPDATE social_tokens
SET 
    access_token = COALESCE(sqlc.narg('access_token'), access_token),
    refresh_token = COALESCE(sqlc.narg('refresh_token'), refresh_token),
    token_type = COALESCE(sqlc.narg('token_type'), token_type),
    expires_at = COALESCE(sqlc.narg('expires_at'), expires_at),
    scope = COALESCE(sqlc.narg('scope'), scope),
    updated_at = NOW()
WHERE social_account_id = sqlc.arg('social_account_id');

-- name: UpdateSocialTokens :exec
UPDATE social_tokens
SET 
    access_token = $2,
    refresh_token = $3,
    expires_at = $4,
    updated_at = NOW()
WHERE social_account_id = $1;

-- name: DeleteSocialToken :exec
DELETE FROM social_tokens
WHERE social_account_id = $1;

-- name: GetExpiringSocialTokens :many
SELECT 
    st.*,
    sa.platform,
    sa.platform_user_id,
    sa.username,
    sa.team_id
FROM social_tokens st
INNER JOIN social_accounts sa ON st.social_account_id = sa.id
WHERE st.expires_at IS NOT NULL 
  AND st.expires_at < NOW() + INTERVAL '7 days'
  AND sa.deleted_at IS NULL
  AND sa.status = 'active'
ORDER BY st.expires_at ASC;

-- name: CountSocialTokensByTeam :one
SELECT COUNT(*)
FROM social_tokens st
INNER JOIN social_accounts sa ON st.social_account_id = sa.id
WHERE sa.team_id = $1
  AND sa.deleted_at IS NULL;