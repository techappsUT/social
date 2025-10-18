-- backend/sql/users.sql
-- ✅ FIXED - User CRUD operations aligned with repository

-- name: CreateUser :one
INSERT INTO users (
    email,
    email_verified,
    password_hash,
    username,
    first_name,
    last_name,
    avatar_url,
    timezone,
    verification_token,                  -- ✅ ADDED
    verification_token_expires_at        -- ✅ ADDED
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
)
RETURNING id, email, email_verified, password_hash, username, first_name, last_name, full_name, 
          avatar_url, timezone, locale, is_active, 
          verification_token, verification_token_expires_at, reset_token, reset_token_expires_at,
          last_login_at, created_at, updated_at, deleted_at;

-- name: GetUserByID :one
SELECT id, email, email_verified, password_hash, username, first_name, last_name, full_name, 
       avatar_url, timezone, locale, is_active,
       verification_token, verification_token_expires_at, reset_token, reset_token_expires_at,
       last_login_at, created_at, updated_at, deleted_at
FROM users
WHERE id = $1 AND deleted_at IS NULL;

-- name: GetUserByEmail :one
SELECT id, email, email_verified, password_hash, username, first_name, last_name, full_name, 
       avatar_url, timezone, locale, is_active,
       verification_token, verification_token_expires_at, reset_token, reset_token_expires_at,
       last_login_at, created_at, updated_at, deleted_at
FROM users
WHERE email = $1 AND deleted_at IS NULL;

-- name: GetUserByUsername :one
SELECT id, email, email_verified, password_hash, username, first_name, last_name, full_name, 
       avatar_url, timezone, locale, is_active,
       verification_token, verification_token_expires_at, reset_token, reset_token_expires_at,
       last_login_at, created_at, updated_at, deleted_at
FROM users
WHERE username = $1 AND deleted_at IS NULL;

-- name: GetUserByIdentifier :one
SELECT id, email, email_verified, password_hash, username, first_name, last_name, full_name, 
       avatar_url, timezone, locale, is_active,
       verification_token, verification_token_expires_at, reset_token, reset_token_expires_at,
       last_login_at, created_at, updated_at, deleted_at
FROM users
WHERE (email = $1 OR username = $1) AND deleted_at IS NULL;

-- ✅ FIX: Change nullable params to non-nullable for required fields
-- name: UpdateUserProfile :one
UPDATE users
SET 
    username = $2,           -- Changed from COALESCE to direct assignment
    first_name = $3,         -- Changed from COALESCE to direct assignment
    last_name = $4,          -- Changed from COALESCE to direct assignment
    avatar_url = COALESCE($5, avatar_url),  -- Keep optional
    timezone = COALESCE($6, timezone),      -- Keep optional
    updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL
RETURNING id, email, email_verified, password_hash, username, first_name, last_name, full_name, 
          avatar_url, timezone, locale, is_active, 
          verification_token, verification_token_expires_at, reset_token, reset_token_expires_at,
          last_login_at, created_at, updated_at, deleted_at;

-- name: UpdateUserPassword :exec
UPDATE users
SET 
    password_hash = $2,
    updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL;

-- name: UpdateUserLastLogin :exec
UPDATE users
SET 
    last_login_at = NOW(),
    updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL;

-- name: SoftDeleteUser :exec
UPDATE users
SET 
    deleted_at = NOW(),
    updated_at = NOW()
WHERE id = $1;

-- name: CheckEmailExists :one
SELECT EXISTS(
    SELECT 1 FROM users 
    WHERE email = $1 AND deleted_at IS NULL
) AS exists;

-- name: CheckUsernameExists :one
SELECT EXISTS(
    SELECT 1 FROM users 
    WHERE username = $1 AND deleted_at IS NULL
) AS exists;

-- ============================================================================
-- ADDITIONAL USEFUL QUERIES
-- ============================================================================

-- name: ListUsers :many
SELECT id, email, email_verified, password_hash, username, first_name, last_name, full_name, 
       avatar_url, timezone, locale, is_active,
       verification_token, verification_token_expires_at, reset_token, reset_token_expires_at,
       last_login_at, created_at, updated_at, deleted_at
FROM users
WHERE deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountUsers :one
SELECT COUNT(*) FROM users WHERE deleted_at IS NULL;

-- name: CountVerifiedUsers :one
SELECT COUNT(*) FROM users WHERE email_verified = TRUE AND deleted_at IS NULL;

-- name: CountUnverifiedUsers :one
SELECT COUNT(*) FROM users WHERE email_verified = FALSE AND deleted_at IS NULL;

-- name: GetUsersByRole :many
-- Note: This requires adding a 'role' column to users table
-- For now, this is a placeholder
SELECT id, email, username FROM users WHERE deleted_at IS NULL;

-- name: UpdateUserStatus :exec
UPDATE users
SET 
    is_active = $2,
    updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL;

-- name: BulkUpdateUserStatus :exec
UPDATE users
SET 
    is_active = $2,
    updated_at = NOW()
WHERE id = ANY($1::uuid[]) AND deleted_at IS NULL;