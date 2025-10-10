-- path: backend/migrations/000005_create_social_tokens.up.sql

CREATE TABLE social_tokens (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    
    -- Platform identification
    platform_type VARCHAR(50) NOT NULL,
    platform_user_id VARCHAR(255) NOT NULL,
    platform_username VARCHAR(255),
    
    -- Encrypted tokens (store as TEXT for encrypted base64 strings)
    access_token TEXT NOT NULL,
    refresh_token TEXT,
    
    -- Token metadata
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    scope TEXT,
    is_valid BOOLEAN DEFAULT true,
    last_validated TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Platform-specific data (JSONB for flexibility)
    extra JSONB,
    
    -- Timestamps
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    
    -- Constraints
    UNIQUE(user_id, platform_type, platform_user_id)
);

CREATE INDEX idx_social_tokens_user_id ON social_tokens(user_id);
CREATE INDEX idx_social_tokens_platform_type ON social_tokens(platform_type);
CREATE INDEX idx_social_tokens_expires_at ON social_tokens(expires_at);
CREATE INDEX idx_social_tokens_is_valid ON social_tokens(is_valid);

-- Trigger to update updated_at
CREATE TRIGGER update_social_tokens_updated_at
    BEFORE UPDATE ON social_tokens
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

COMMENT ON TABLE social_tokens IS 'Stores encrypted OAuth tokens for social platform integrations';
COMMENT ON COLUMN social_tokens.access_token IS 'AES-256-GCM encrypted access token';
COMMENT ON COLUMN social_tokens.refresh_token IS 'AES-256-GCM encrypted refresh token';