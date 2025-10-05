-- path: backend/sql/social_accounts.sql

-- name: LinkSocialAccount :one
INSERT INTO social_accounts (
    team_id,
    platform,
    platform_user_id,
    username,
    display_name,
    avatar_url,
    profile_url,
    account_type,
    status,
    metadata,
    connected_by
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11
)
RETURNING *;

-- name: GetSocialAccountByID :one
SELECT * FROM social_accounts
WHERE id = $1 AND deleted_at IS NULL;

-- name: GetSocialAccountWithToken :one
SELECT 
    sa.*,
    st.access_token,
    st.refresh_token,
    st.expires_at as token_expires_at
FROM social_accounts sa
LEFT JOIN social_tokens st ON sa.id = st.social_account_id
WHERE sa.id = $1 AND sa.deleted_at IS NULL;

-- name: ListSocialAccountsByTeam :many
SELECT * FROM social_accounts
WHERE team_id = $1 AND deleted_at IS NULL
ORDER BY platform ASC, created_at ASC;

-- name: ListSocialAccountsByPlatform :many
SELECT * FROM social_accounts
WHERE team_id = $1 
  AND platform = $2 
  AND deleted_at IS NULL
ORDER BY created_at ASC;

-- name: UpdateSocialAccountStatus :exec
UPDATE social_accounts
SET 
    status = $2,
    metadata = COALESCE(sqlc.narg('metadata'), metadata),
    updated_at = NOW()
WHERE id = $1;

-- name: UpdateSocialAccountMetadata :exec
UPDATE social_accounts
SET 
    username = COALESCE(sqlc.narg('username'), username),
    display_name = COALESCE(sqlc.narg('display_name'), display_name),
    avatar_url = COALESCE(sqlc.narg('avatar_url'), avatar_url),
    metadata = COALESCE(sqlc.narg('metadata'), metadata),
    last_synced_at = NOW(),
    updated_at = NOW()
WHERE id = sqlc.arg('id');

-- name: SoftDeleteSocialAccount :exec
UPDATE social_accounts
SET 
    deleted_at = NOW(),
    updated_at = NOW()
WHERE id = $1;