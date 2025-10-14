-- backend/migrations/000002_add_auth_tokens_to_users.down.sql

-- Drop indexes
DROP INDEX IF EXISTS idx_users_verification_token;
DROP INDEX IF EXISTS idx_users_reset_token;

-- Drop columns
ALTER TABLE users 
    DROP COLUMN IF EXISTS verification_token,
    DROP COLUMN IF EXISTS verification_token_expires_at,
    DROP COLUMN IF EXISTS reset_token,
    DROP COLUMN IF EXISTS reset_token_expires_at;