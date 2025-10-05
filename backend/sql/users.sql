-- path: backend/sql/users.sql

-- name: CreateUser :one
INSERT INTO users (
    email,
    email_verified,
    password_hash,
    full_name,
    avatar_url,
    timezone
) VALUES (
    $1, $2, $3, $4, $5, $6
)
RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users
WHERE id = $1 AND deleted_at IS NULL;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1 AND deleted_at IS NULL;

-- name: UpdateUserProfile :one
UPDATE users
SET 
    full_name = COALESCE(sqlc.narg('full_name'), full_name),
    avatar_url = COALESCE(sqlc.narg('avatar_url'), avatar_url),
    timezone = COALESCE(sqlc.narg('timezone'), timezone),
    updated_at = NOW()
WHERE id = sqlc.arg('id') AND deleted_at IS NULL
RETURNING *;

-- name: UpdateUserPassword :exec
UPDATE users
SET 
    password_hash = $2,
    updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL;

-- name: UpdateUserLastLogin :exec
UPDATE users
SET last_login_at = NOW()
WHERE id = $1;

-- name: MarkUserEmailVerified :exec
UPDATE users
SET 
    email_verified = true,
    updated_at = NOW()
WHERE id = $1;

-- name: SoftDeleteUser :exec
UPDATE users
SET 
    deleted_at = NOW(),
    updated_at = NOW()
WHERE id = $1;