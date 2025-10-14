-- path: backend/migrations/20240101000001_initial_schema.up.sql

-- ============================================================================
-- EXTENSIONS
-- ============================================================================

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm"; -- for text search


-- ============================================================================
-- HELPER FUNCTIONS
-- ============================================================================

-- Updated at trigger function
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';


-- ============================================================================
-- USERS & AUTHENTICATION
-- ============================================================================

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    email_verified BOOLEAN DEFAULT FALSE,
    password_hash VARCHAR(255), -- nullable for OAuth-only users
    username VARCHAR(30) UNIQUE NOT NULL,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    full_name VARCHAR(255), -- deprecated, for backward compatibility
    avatar_url TEXT,
    timezone VARCHAR(50) DEFAULT 'UTC',
    locale VARCHAR(10) DEFAULT 'en',
    is_active BOOLEAN DEFAULT TRUE,
    last_login_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_users_email ON users(email) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_username ON users(username) WHERE deleted_at IS NULL;
CREATE INDEX idx_users_deleted_at ON users(deleted_at) WHERE deleted_at IS NOT NULL;

-- Add check constraint for username format
ALTER TABLE users ADD CONSTRAINT users_username_format 
    CHECK (username ~ '^[a-z0-9_-]{3,30}$');

CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Comments
COMMENT ON COLUMN users.username IS 'Unique username for user identification (3-30 chars, lowercase alphanumeric with _ and -)';
COMMENT ON COLUMN users.first_name IS 'User first name (required)';
COMMENT ON COLUMN users.last_name IS 'User last name (optional but recommended)';
COMMENT ON COLUMN users.full_name IS 'DEPRECATED: Use first_name and last_name instead';


-- ============================================================================
-- TEAMS & MULTI-TENANCY
-- ============================================================================

CREATE TABLE teams (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) UNIQUE NOT NULL,
    avatar_url TEXT,
    settings JSONB DEFAULT '{}',
    is_active BOOLEAN DEFAULT TRUE,
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_teams_slug ON teams(slug) WHERE deleted_at IS NULL;
CREATE INDEX idx_teams_created_by ON teams(created_by);

CREATE TRIGGER update_teams_updated_at BEFORE UPDATE ON teams
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();


-- ============================================================================
-- ROLES & PERMISSIONS
-- ============================================================================

CREATE TABLE roles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(50) UNIQUE NOT NULL,
    description TEXT,
    permissions JSONB DEFAULT '[]',
    is_system BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TRIGGER update_roles_updated_at BEFORE UPDATE ON roles
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Insert default roles
INSERT INTO roles (name, description, permissions, is_system) VALUES
    ('owner', 'Team owner with full access', '["*"]', TRUE),
    ('admin', 'Administrator with management access', '["team.manage", "users.manage", "posts.manage", "analytics.view"]', TRUE),
    ('member', 'Regular member who can create and schedule posts', '["posts.create", "posts.manage_own", "analytics.view"]', TRUE),
    ('viewer', 'Read-only access to posts and analytics', '["posts.view", "analytics.view"]', TRUE);


-- ============================================================================
-- TEAM MEMBERSHIPS
-- ============================================================================

CREATE TABLE team_memberships (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    team_id UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE RESTRICT,
    invited_by UUID REFERENCES users(id) ON DELETE SET NULL,
    invitation_token VARCHAR(100) UNIQUE,
    invitation_accepted_at TIMESTAMPTZ,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    UNIQUE(team_id, user_id, deleted_at)
);

CREATE INDEX idx_team_memberships_team_id ON team_memberships(team_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_team_memberships_user_id ON team_memberships(user_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_team_memberships_invitation_token ON team_memberships(invitation_token) WHERE invitation_token IS NOT NULL;

CREATE TRIGGER update_team_memberships_updated_at BEFORE UPDATE ON team_memberships
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();


-- ============================================================================
-- REFRESH TOKENS (JWT)
-- ============================================================================

CREATE TABLE refresh_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX idx_refresh_tokens_token_hash ON refresh_tokens(token_hash) WHERE revoked = FALSE;
CREATE INDEX idx_refresh_tokens_expires_at ON refresh_tokens(expires_at) WHERE revoked = FALSE;


-- ============================================================================
-- SOCIAL ACCOUNTS & INTEGRATIONS
-- ============================================================================

CREATE TYPE social_platform AS ENUM (
    'twitter',
    'facebook',
    'instagram',
    'linkedin',
    'tiktok',
    'youtube',
    'pinterest',
    'threads'
);

CREATE TYPE social_account_status AS ENUM (
    'active',
    'expired',
    'revoked',
    'error'
);

CREATE TABLE social_accounts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    team_id UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    platform social_platform NOT NULL,
    platform_user_id VARCHAR(255) NOT NULL,
    username VARCHAR(255),
    display_name VARCHAR(255),
    avatar_url TEXT,
    profile_url TEXT,
    account_type VARCHAR(50),
    status social_account_status DEFAULT 'active',
    metadata JSONB DEFAULT '{}',
    connected_by UUID REFERENCES users(id) ON DELETE SET NULL,
    connected_at TIMESTAMPTZ DEFAULT NOW(),
    last_synced_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    UNIQUE(team_id, platform, platform_user_id, deleted_at)
);

CREATE INDEX idx_social_accounts_team_id ON social_accounts(team_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_social_accounts_platform ON social_accounts(platform);
CREATE INDEX idx_social_accounts_status ON social_accounts(status);

CREATE TRIGGER update_social_accounts_updated_at BEFORE UPDATE ON social_accounts
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();


-- ============================================================================
-- SOCIAL TOKENS (OAuth credentials)
-- ============================================================================

CREATE TABLE social_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    social_account_id UUID UNIQUE NOT NULL REFERENCES social_accounts(id) ON DELETE CASCADE,
    access_token TEXT NOT NULL,
    refresh_token TEXT,
    token_type VARCHAR(50) DEFAULT 'Bearer',
    expires_at TIMESTAMPTZ,
    scope TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_social_tokens_expires_at ON social_tokens(expires_at);

CREATE TRIGGER update_social_tokens_updated_at BEFORE UPDATE ON social_tokens
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();


-- ============================================================================
-- SCHEDULED POSTS
-- ============================================================================

CREATE TYPE post_status AS ENUM (
    'draft',
    'scheduled',
    'queued',
    'processing',
    'published',
    'failed',
    'cancelled'
);

CREATE TABLE scheduled_posts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    team_id UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    created_by UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    social_account_id UUID NOT NULL REFERENCES social_accounts(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    content_html TEXT,
    shortened_links JSONB DEFAULT '[]',
    status post_status DEFAULT 'draft',
    scheduled_at TIMESTAMPTZ,
    published_at TIMESTAMPTZ,
    platform_specific_options JSONB DEFAULT '{}',
    error_message TEXT,
    retry_count INTEGER DEFAULT 0,
    max_retries INTEGER DEFAULT 3,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_scheduled_posts_team_id ON scheduled_posts(team_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_scheduled_posts_status ON scheduled_posts(status);
CREATE INDEX idx_scheduled_posts_scheduled_at ON scheduled_posts(scheduled_at);
CREATE INDEX idx_scheduled_posts_created_by ON scheduled_posts(created_by);

CREATE TRIGGER update_scheduled_posts_updated_at BEFORE UPDATE ON scheduled_posts
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();


-- ============================================================================
-- POST ATTACHMENTS
-- ============================================================================

CREATE TYPE attachment_type AS ENUM ('image', 'video', 'gif', 'document');

CREATE TABLE post_attachments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    scheduled_post_id UUID NOT NULL REFERENCES scheduled_posts(id) ON DELETE CASCADE,
    type attachment_type NOT NULL,
    url TEXT NOT NULL,
    thumbnail_url TEXT,
    file_size BIGINT,
    mime_type VARCHAR(100),
    width INTEGER,
    height INTEGER,
    duration INTEGER,
    alt_text TEXT,
    upload_order INTEGER DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_post_attachments_post_id ON post_attachments(scheduled_post_id);


-- ============================================================================
-- PUBLISHED POSTS (Archive)
-- ============================================================================

CREATE TABLE posts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    scheduled_post_id UUID UNIQUE REFERENCES scheduled_posts(id) ON DELETE SET NULL,
    team_id UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    social_account_id UUID NOT NULL REFERENCES social_accounts(id) ON DELETE CASCADE,
    platform_post_id VARCHAR(255),
    platform_post_url TEXT,
    content TEXT NOT NULL,
    published_at TIMESTAMPTZ DEFAULT NOW(),
    metrics JSONB DEFAULT '{}',
    last_metrics_fetch_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_posts_team_id ON posts(team_id);
CREATE INDEX idx_posts_social_account_id ON posts(social_account_id);
CREATE INDEX idx_posts_platform_post_id ON posts(platform_post_id);
CREATE INDEX idx_posts_published_at ON posts(published_at);

CREATE TRIGGER update_posts_updated_at BEFORE UPDATE ON posts
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();


-- ============================================================================
-- POST QUEUE (Background Job Queue)
-- ============================================================================

CREATE TYPE queue_status AS ENUM ('pending', 'processing', 'completed', 'failed');

CREATE TABLE post_queue (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    scheduled_post_id UUID NOT NULL REFERENCES scheduled_posts(id) ON DELETE CASCADE,
    status queue_status DEFAULT 'pending',
    priority INTEGER DEFAULT 0,
    attempts INTEGER DEFAULT 0,
    max_attempts INTEGER DEFAULT 3,
    error TEXT,
    scheduled_for TIMESTAMPTZ NOT NULL,
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_post_queue_status ON post_queue(status);
CREATE INDEX idx_post_queue_scheduled_for ON post_queue(scheduled_for);
CREATE INDEX idx_post_queue_priority ON post_queue(priority DESC);

CREATE TRIGGER update_post_queue_updated_at BEFORE UPDATE ON post_queue
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();


-- ============================================================================
-- ANALYTICS EVENTS
-- ============================================================================

CREATE TYPE event_type AS ENUM (
    'impression',
    'click',
    'like',
    'share',
    'comment',
    'retweet',
    'reply',
    'view',
    'save'
);

CREATE TABLE analytics_events (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    post_id UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    event_type event_type NOT NULL,
    event_value INTEGER DEFAULT 1,
    event_metadata JSONB DEFAULT '{}',
    recorded_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_analytics_events_post_id ON analytics_events(post_id);
CREATE INDEX idx_analytics_events_type ON analytics_events(event_type);
CREATE INDEX idx_analytics_events_recorded_at ON analytics_events(recorded_at);


-- ============================================================================
-- BILLING & SUBSCRIPTIONS
-- ============================================================================

CREATE TYPE plan_interval AS ENUM ('month', 'year');

CREATE TABLE plans (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) UNIQUE NOT NULL,
    slug VARCHAR(50) UNIQUE NOT NULL,
    description TEXT,
    price_monthly DECIMAL(10, 2) NOT NULL,
    price_yearly DECIMAL(10, 2) NOT NULL,
    features JSONB DEFAULT '{}',
    limits JSONB DEFAULT '{}',
    is_active BOOLEAN DEFAULT TRUE,
    stripe_price_id_monthly VARCHAR(255),
    stripe_price_id_yearly VARCHAR(255),
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TRIGGER update_plans_updated_at BEFORE UPDATE ON plans
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Insert default plans
INSERT INTO plans (name, slug, description, price_monthly, price_yearly, features, limits, is_active) VALUES
    ('Free', 'free', 'Perfect for getting started', 0.00, 0.00,
     '["3 social accounts", "10 scheduled posts", "Basic analytics"]'::jsonb,
     '{"social_accounts": 3, "scheduled_posts": 10, "team_members": 1}'::jsonb,
     TRUE),
    ('Pro', 'pro', 'For growing businesses', 19.99, 199.00,
     '["Unlimited accounts", "Unlimited posts", "Advanced analytics", "Team collaboration"]'::jsonb,
     '{"social_accounts": -1, "scheduled_posts": -1, "team_members": 5}'::jsonb,
     TRUE),
    ('Enterprise', 'enterprise', 'For large organizations', 99.99, 999.00,
     '["Everything in Pro", "Priority support", "Custom integrations", "Dedicated account manager"]'::jsonb,
     '{"social_accounts": -1, "scheduled_posts": -1, "team_members": -1}'::jsonb,
     TRUE);


CREATE TYPE subscription_status AS ENUM (
    'active',
    'trialing',
    'past_due',
    'canceled',
    'unpaid'
);

CREATE TABLE subscriptions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    team_id UUID UNIQUE NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    plan_id UUID NOT NULL REFERENCES plans(id) ON DELETE RESTRICT,
    status subscription_status DEFAULT 'active',
    interval plan_interval NOT NULL,
    stripe_subscription_id VARCHAR(255) UNIQUE,
    stripe_customer_id VARCHAR(255),
    current_period_start TIMESTAMPTZ,
    current_period_end TIMESTAMPTZ,
    cancel_at_period_end BOOLEAN DEFAULT FALSE,
    canceled_at TIMESTAMPTZ,
    trial_start TIMESTAMPTZ,
    trial_end TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_subscriptions_team_id ON subscriptions(team_id);
CREATE INDEX idx_subscriptions_status ON subscriptions(status);
CREATE INDEX idx_subscriptions_stripe_customer_id ON subscriptions(stripe_customer_id);

CREATE TRIGGER update_subscriptions_updated_at BEFORE UPDATE ON subscriptions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();


CREATE TYPE invoice_status AS ENUM ('draft', 'open', 'paid', 'void', 'uncollectible');

CREATE TABLE invoices (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    subscription_id UUID NOT NULL REFERENCES subscriptions(id) ON DELETE CASCADE,
    stripe_invoice_id VARCHAR(255) UNIQUE NOT NULL,
    amount_due DECIMAL(10, 2) NOT NULL,
    amount_paid DECIMAL(10, 2) DEFAULT 0.00,
    currency VARCHAR(3) DEFAULT 'USD',
    status invoice_status DEFAULT 'draft',
    invoice_pdf TEXT,
    hosted_invoice_url TEXT,
    due_date TIMESTAMPTZ,
    paid_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_invoices_subscription_id ON invoices(subscription_id);
CREATE INDEX idx_invoices_status ON invoices(status);

CREATE TRIGGER update_invoices_updated_at BEFORE UPDATE ON invoices
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();


-- ============================================================================
-- WEBHOOKS & INTEGRATIONS
-- ============================================================================

CREATE TYPE webhook_source AS ENUM ('stripe', 'social_platform', 'internal');

CREATE TABLE webhooks_log (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    source webhook_source NOT NULL,
    event_type VARCHAR(100) NOT NULL,
    payload JSONB NOT NULL,
    processed BOOLEAN DEFAULT FALSE,
    processed_at TIMESTAMPTZ,
    error TEXT,
    received_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_webhooks_log_source ON webhooks_log(source);
CREATE INDEX idx_webhooks_log_processed ON webhooks_log(processed);
CREATE INDEX idx_webhooks_log_received_at ON webhooks_log(received_at);


-- ============================================================================
-- BACKGROUND JOBS
-- ============================================================================

CREATE TYPE job_status AS ENUM ('pending', 'running', 'completed', 'failed');

CREATE TABLE job_runs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    job_name VARCHAR(100) NOT NULL,
    status job_status DEFAULT 'pending',
    payload JSONB DEFAULT '{}',
    result JSONB,
    error TEXT,
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_job_runs_job_name ON job_runs(job_name);
CREATE INDEX idx_job_runs_status ON job_runs(status);
CREATE INDEX idx_job_runs_created_at ON job_runs(created_at);

CREATE TRIGGER update_job_runs_updated_at BEFORE UPDATE ON job_runs
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();


-- ============================================================================
-- COMMENTS
-- ============================================================================

COMMENT ON TABLE users IS 'User accounts with authentication';
COMMENT ON TABLE teams IS 'Organizations/teams for multi-tenancy';
COMMENT ON TABLE roles IS 'Role-based access control';
COMMENT ON TABLE team_memberships IS 'User membership in teams with roles';
COMMENT ON TABLE refresh_tokens IS 'JWT refresh tokens for session management';
COMMENT ON TABLE social_accounts IS 'Connected social media accounts';
COMMENT ON TABLE social_tokens IS 'OAuth tokens for social platforms (encrypted)';
COMMENT ON TABLE scheduled_posts IS 'Posts scheduled for future publishing';
COMMENT ON TABLE post_attachments IS 'Media attachments for posts';
COMMENT ON TABLE posts IS 'Archive of published posts';
COMMENT ON TABLE post_queue IS 'Background job queue for post publishing';
COMMENT ON TABLE analytics_events IS 'Social media engagement events';
COMMENT ON TABLE plans IS 'Subscription plans';
COMMENT ON TABLE subscriptions IS 'Team subscriptions';
COMMENT ON TABLE invoices IS 'Billing invoices';
COMMENT ON TABLE webhooks_log IS 'Webhook event log';
COMMENT ON TABLE job_runs IS 'Background job execution log';-- backend/migrations/20240101000002_add_auth_tokens_to_users.up.sql

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