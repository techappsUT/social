-- backend/migrations/20240101000002_add_auth_tokens_to_users.up.sql

-- Add email verification and password reset token columns to users table
ALTER TABLE users
    ADD COLUMN verification_token VARCHAR(255),
    ADD COLUMN verification_token_expires_at TIMESTAMPTZ,
    ADD COLUMN reset_token VARCHAR(255),
    ADD COLUMN reset_token_expires_at TIMESTAMPTZ;

-- Add indexes for token lookups (simplified - no time-based predicate)
CREATE INDEX idx_users_verification_token ON users(verification_token)
    WHERE verification_token IS NOT NULL;

CREATE INDEX idx_users_reset_token ON users(reset_token)
    WHERE reset_token IS NOT NULL;

-- Comments
COMMENT ON COLUMN users.verification_token IS 'Email verification token (expires in 24 hours)';
COMMENT ON COLUMN users.reset_token IS 'Password reset token (expires in 1 hour)';