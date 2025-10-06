-- path: backend/sql/social_tokens.sql

-- name: UpsertSocialToken :one
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
ON CONFLICT (social_account_id) 
DO UPDATE SET
    access_token = EXCLUDED.access_token,
    refresh_token = EXCLUDED.refresh_token,
    token_type = EXCLUDED.token_type,
    expires_at = EXCLUDED.expires_at,
    scope = EXCLUDED.scope,
    updated_at = NOW()
RETURNING *;

-- name: GetSocialToken :one
SELECT * FROM social_tokens
WHERE social_account_id = $1;

-- name: ListExpiringTokens :many
SELECT st.*, sa.team_id, sa.platform
FROM social_tokens st
INNER JOIN social_accounts sa ON st.social_account_id = sa.id
WHERE st.expires_at < $1
  AND sa.status = 'active'
  AND sa.deleted_at IS NULL
ORDER BY st.expires_at ASC;