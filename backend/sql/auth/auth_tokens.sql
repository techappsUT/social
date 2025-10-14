-- backend/sql/auth/auth_tokens.sql

-- ============================================================================
-- EMAIL VERIFICATION TOKENS
-- ============================================================================

-- name: SetVerificationToken :exec
UPDATE users
SET verification_token = $2,
    verification_token_expires_at = $3,
    updated_at = NOW()
WHERE id = $1;

-- name: GetUserByVerificationToken :one
SELECT id, email, email_verified, password_hash, username, first_name, last_name, full_name, 
       avatar_url, timezone, locale, is_active,
       verification_token, verification_token_expires_at, reset_token, reset_token_expires_at,
       last_login_at, created_at, updated_at, deleted_at
FROM users
WHERE verification_token = $1 
  AND verification_token_expires_at > NOW()
  AND deleted_at IS NULL
LIMIT 1;

-- name: ClearVerificationToken :exec
UPDATE users
SET verification_token = NULL,
    verification_token_expires_at = NULL,
    email_verified = TRUE,
    updated_at = NOW()
WHERE id = $1;

-- ============================================================================
-- PASSWORD RESET TOKENS
-- ============================================================================

-- name: SetResetToken :exec
UPDATE users
SET reset_token = $2,
    reset_token_expires_at = $3,
    updated_at = NOW()
WHERE id = $1;

-- name: GetUserByResetToken :one
SELECT id, email, email_verified, password_hash, username, first_name, last_name, full_name, 
       avatar_url, timezone, locale, is_active,
       verification_token, verification_token_expires_at, reset_token, reset_token_expires_at,
       last_login_at, created_at, updated_at, deleted_at
FROM users
WHERE reset_token = $1 
  AND reset_token_expires_at > NOW()
  AND deleted_at IS NULL
LIMIT 1;

-- name: ClearResetToken :exec
UPDATE users
SET reset_token = NULL,
    reset_token_expires_at = NULL,
    updated_at = NOW()
WHERE id = $1;

-- ============================================================================
-- TOKEN CLEANUP
-- ============================================================================

-- name: ClearExpiredVerificationTokens :exec
UPDATE users
SET verification_token = NULL,
    verification_token_expires_at = NULL
WHERE verification_token_expires_at < NOW();

-- name: ClearExpiredResetTokens :exec
UPDATE users
SET reset_token = NULL,
    reset_token_expires_at = NULL
WHERE reset_token_expires_at < NOW();