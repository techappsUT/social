-- backend/sql/users.sql
-- User CRUD operations

-- name: CreateUser :one
INSERT INTO users (
    email,
    email_verified,
    password_hash,
    username,
    first_name,
    last_name,
    avatar_url,
    timezone
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
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

-- name: UpdateUserProfile :one
UPDATE users
SET 
    username = COALESCE($1, username),
    first_name = COALESCE($2, first_name),
    last_name = COALESCE($3, last_name),
    avatar_url = COALESCE($4, avatar_url),
    timezone = COALESCE($5, timezone),
    updated_at = NOW()
WHERE id = $6 AND deleted_at IS NULL
RETURNING id, email, email_verified, password_hash, username, first_name, last_name, full_name, 
          avatar_url, timezone, locale, is_active,
          verification_token, verification_token_expires_at, reset_token, reset_token_expires_at,
          last_login_at, created_at, updated_at, deleted_at;

-- name: UpdateUserPassword :exec
UPDATE users
SET password_hash = $2,
    updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL;

-- name: UpdateUserLastLogin :exec
UPDATE users
SET last_login_at = NOW(),
    updated_at = NOW()
WHERE id = $1;

-- name: SoftDeleteUser :exec
UPDATE users
SET deleted_at = NOW()
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