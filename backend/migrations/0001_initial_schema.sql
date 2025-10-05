-- path: backend/migrations/0001_initial_schema.sql

-- ============================================================================
-- EXTENSIONS
-- ============================================================================

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pg_trgm"; -- for text search


-- ============================================================================
-- USERS & AUTHENTICATION
-- ============================================================================

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email VARCHAR(255) UNIQUE NOT NULL,
    email_verified BOOLEAN DEFAULT FALSE,
    password_hash VARCHAR(255), -- nullable for OAuth-only users
    full_name VARCHAR(255),
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
CREATE INDEX idx_users_deleted_at ON users(deleted_at) WHERE deleted_at IS NOT NULL;


-- ============================================================================
-- TEAMS & MULTI-TENANCY
-- ============================================================================

CREATE TABLE teams (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(100) UNIQUE NOT NULL,
    avatar_url TEXT,
    settings JSONB DEFAULT '{}', -- team preferences, branding, etc.
    is_active BOOLEAN DEFAULT TRUE,
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_teams_slug ON teams(slug) WHERE deleted_at IS NULL;
CREATE INDEX idx_teams_created_by ON teams(created_by);


-- ============================================================================
-- ROLES & PERMISSIONS
-- ============================================================================

CREATE TABLE roles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(50) UNIQUE NOT NULL, -- 'owner', 'admin', 'member', 'viewer'
    description TEXT,
    permissions JSONB DEFAULT '[]', -- array of permission strings
    is_system BOOLEAN DEFAULT FALSE, -- system roles cannot be deleted
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

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
    UNIQUE(team_id, user_id, deleted_at) -- prevent duplicate active memberships
);

CREATE INDEX idx_team_memberships_team_id ON team_memberships(team_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_team_memberships_user_id ON team_memberships(user_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_team_memberships_invitation_token ON team_memberships(invitation_token) WHERE invitation_token IS NOT NULL;


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
    platform_user_id VARCHAR(255) NOT NULL, -- external platform's user ID
    username VARCHAR(255),
    display_name VARCHAR(255),
    avatar_url TEXT,
    profile_url TEXT,
    account_type VARCHAR(50), -- 'personal', 'business', 'page', etc.
    status social_account_status DEFAULT 'active',
    metadata JSONB DEFAULT '{}', -- platform-specific data (follower_count, etc.)
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


-- ============================================================================
-- SOCIAL TOKENS (OAuth credentials)
-- ============================================================================

CREATE TABLE social_tokens (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    social_account_id UUID UNIQUE NOT NULL REFERENCES social_accounts(id) ON DELETE CASCADE,
    access_token TEXT NOT NULL, -- encrypted in application
    refresh_token TEXT, -- encrypted in application
    token_type VARCHAR(50) DEFAULT 'Bearer',
    expires_at TIMESTAMPTZ,
    scope TEXT, -- comma-separated scopes
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_social_tokens_expires_at ON social_tokens(expires_at);


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
    
    -- Content
    content TEXT NOT NULL,
    content_html TEXT, -- for platforms supporting rich text
    shortened_links JSONB DEFAULT '[]', -- array of {original, shortened}
    
    -- Scheduling
    status post_status DEFAULT 'draft',
    scheduled_at TIMESTAMPTZ,
    published_at TIMESTAMPTZ,
    
    -- Metadata
    platform_specific_options JSONB DEFAULT '{}', -- thread settings, first_comment, etc.
    retry_count INT DEFAULT 0,
    error_message TEXT,
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_scheduled_posts_team_id ON scheduled_posts(team_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_scheduled_posts_status ON scheduled_posts(status);
CREATE INDEX idx_scheduled_posts_scheduled_at ON scheduled_posts(scheduled_at) WHERE status IN ('scheduled', 'queued');
CREATE INDEX idx_scheduled_posts_created_by ON scheduled_posts(created_by);
CREATE INDEX idx_scheduled_posts_social_account_id ON scheduled_posts(social_account_id);


-- ============================================================================
-- POST ATTACHMENTS
-- ============================================================================

CREATE TYPE attachment_type AS ENUM (
    'image',
    'video',
    'gif',
    'document',
    'link_preview'
);

CREATE TABLE post_attachments (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    scheduled_post_id UUID NOT NULL REFERENCES scheduled_posts(id) ON DELETE CASCADE,
    type attachment_type NOT NULL,
    url TEXT NOT NULL, -- S3/CDN URL
    thumbnail_url TEXT,
    file_size BIGINT, -- bytes
    mime_type VARCHAR(100),
    width INT,
    height INT,
    duration INT, -- seconds for videos
    alt_text TEXT,
    display_order INT DEFAULT 0,
    upload_metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_post_attachments_scheduled_post_id ON post_attachments(scheduled_post_id);
CREATE INDEX idx_post_attachments_display_order ON post_attachments(scheduled_post_id, display_order);


-- ============================================================================
-- POSTS (Published posts)
-- ============================================================================

CREATE TABLE posts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    scheduled_post_id UUID UNIQUE NOT NULL REFERENCES scheduled_posts(id) ON DELETE CASCADE,
    team_id UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    social_account_id UUID NOT NULL REFERENCES social_accounts(id) ON DELETE CASCADE,
    
    platform_post_id VARCHAR(255) NOT NULL, -- external platform's post ID
    platform_post_url TEXT,
    
    content TEXT,
    published_at TIMESTAMPTZ NOT NULL,
    
    -- Cached analytics (updated periodically)
    impressions BIGINT DEFAULT 0,
    engagements BIGINT DEFAULT 0,
    clicks BIGINT DEFAULT 0,
    likes BIGINT DEFAULT 0,
    shares BIGINT DEFAULT 0,
    comments BIGINT DEFAULT 0,
    
    last_analytics_fetch_at TIMESTAMPTZ,
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    
    UNIQUE(social_account_id, platform_post_id)
);

CREATE INDEX idx_posts_team_id ON posts(team_id);
CREATE INDEX idx_posts_social_account_id ON posts(social_account_id);
CREATE INDEX idx_posts_published_at ON posts(published_at DESC);
CREATE INDEX idx_posts_platform_post_id ON posts(platform_post_id);


-- ============================================================================
-- POST QUEUE (Processing queue)
-- ============================================================================

CREATE TYPE queue_status AS ENUM (
    'pending',
    'processing',
    'completed',
    'failed',
    'retrying'
);

CREATE TABLE post_queue (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    scheduled_post_id UUID NOT NULL REFERENCES scheduled_posts(id) ON DELETE CASCADE,
    status queue_status DEFAULT 'pending',
    priority INT DEFAULT 5, -- 1-10, lower = higher priority
    scheduled_for TIMESTAMPTZ NOT NULL,
    attempt_count INT DEFAULT 0,
    max_attempts INT DEFAULT 3,
    last_error TEXT,
    worker_id VARCHAR(100), -- ID of worker processing this job
    locked_at TIMESTAMPTZ,
    lock_expires_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_post_queue_status ON post_queue(status);
CREATE INDEX idx_post_queue_scheduled_for ON post_queue(scheduled_for) WHERE status = 'pending';
CREATE INDEX idx_post_queue_lock_expires_at ON post_queue(lock_expires_at) WHERE status = 'processing';


-- ============================================================================
-- ANALYTICS EVENTS
-- ============================================================================

CREATE TYPE event_type AS ENUM (
    'impression',
    'click',
    'like',
    'share',
    'comment',
    'follow',
    'engagement'
);

CREATE TABLE analytics_events (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    post_id UUID NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    team_id UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    
    event_type event_type NOT NULL,
    event_value INT DEFAULT 1, -- count or value
    
    -- Dimensions
    platform social_platform NOT NULL,
    country VARCHAR(2), -- ISO country code
    device_type VARCHAR(50), -- mobile, desktop, tablet
    referrer TEXT,
    
    event_metadata JSONB DEFAULT '{}',
    event_timestamp TIMESTAMPTZ NOT NULL,
    
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Partitioning by month recommended for large scale
CREATE INDEX idx_analytics_events_post_id ON analytics_events(post_id);
CREATE INDEX idx_analytics_events_team_id ON analytics_events(team_id);
CREATE INDEX idx_analytics_events_timestamp ON analytics_events(event_timestamp DESC);
CREATE INDEX idx_analytics_events_type ON analytics_events(event_type);


-- ============================================================================
-- SUBSCRIPTION PLANS
-- ============================================================================

CREATE TYPE plan_interval AS ENUM (
    'month',
    'year'
);

CREATE TABLE plans (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) NOT NULL,
    slug VARCHAR(50) UNIQUE NOT NULL,
    description TEXT,
    
    -- Pricing
    price_cents INT NOT NULL, -- in cents (e.g., 2900 = $29.00)
    currency VARCHAR(3) DEFAULT 'USD',
    interval plan_interval NOT NULL,
    
    -- Limits
    features JSONB DEFAULT '{}', -- {social_accounts: 10, posts_per_month: 100, team_members: 5}
    
    -- Stripe
    stripe_price_id VARCHAR(255),
    stripe_product_id VARCHAR(255),
    
    is_active BOOLEAN DEFAULT TRUE,
    is_popular BOOLEAN DEFAULT FALSE,
    display_order INT DEFAULT 0,
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_plans_slug ON plans(slug) WHERE is_active = TRUE;
CREATE INDEX idx_plans_active ON plans(is_active, display_order);

-- Insert default plans
INSERT INTO plans (name, slug, price_cents, currency, interval, features, is_popular, display_order) VALUES
    ('Free', 'free', 0, 'USD', 'month', '{"social_accounts": 3, "posts_per_month": 10, "team_members": 1}', FALSE, 1),
    ('Pro', 'pro', 1500, 'USD', 'month', '{"social_accounts": 10, "posts_per_month": 100, "team_members": 5, "analytics": true}', TRUE, 2),
    ('Business', 'business', 5000, 'USD', 'month', '{"social_accounts": 50, "posts_per_month": -1, "team_members": 20, "analytics": true, "priority_support": true}', FALSE, 3);


-- ============================================================================
-- SUBSCRIPTIONS
-- ============================================================================

CREATE TYPE subscription_status AS ENUM (
    'trialing',
    'active',
    'past_due',
    'canceled',
    'incomplete',
    'incomplete_expired',
    'paused'
);

CREATE TABLE subscriptions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    team_id UUID UNIQUE NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    plan_id UUID NOT NULL REFERENCES plans(id) ON DELETE RESTRICT,
    
    status subscription_status DEFAULT 'active',
    
    -- Stripe
    stripe_subscription_id VARCHAR(255) UNIQUE,
    stripe_customer_id VARCHAR(255),
    
    -- Billing
    current_period_start TIMESTAMPTZ,
    current_period_end TIMESTAMPTZ,
    cancel_at_period_end BOOLEAN DEFAULT FALSE,
    canceled_at TIMESTAMPTZ,
    ended_at TIMESTAMPTZ,
    
    -- Trial
    trial_start TIMESTAMPTZ,
    trial_end TIMESTAMPTZ,
    
    metadata JSONB DEFAULT '{}',
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_subscriptions_team_id ON subscriptions(team_id);
CREATE INDEX idx_subscriptions_status ON subscriptions(status);
CREATE INDEX idx_subscriptions_stripe_customer_id ON subscriptions(stripe_customer_id);
CREATE INDEX idx_subscriptions_current_period_end ON subscriptions(current_period_end) WHERE status = 'active';


-- ============================================================================
-- INVOICES
-- ============================================================================

CREATE TYPE invoice_status AS ENUM (
    'draft',
    'open',
    'paid',
    'void',
    'uncollectible'
);

CREATE TABLE invoices (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    subscription_id UUID NOT NULL REFERENCES subscriptions(id) ON DELETE CASCADE,
    team_id UUID NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    
    stripe_invoice_id VARCHAR(255) UNIQUE,
    invoice_number VARCHAR(100),
    
    status invoice_status DEFAULT 'draft',
    
    -- Amounts in cents
    subtotal BIGINT NOT NULL,
    tax BIGINT DEFAULT 0,
    total BIGINT NOT NULL,
    amount_paid BIGINT DEFAULT 0,
    amount_due BIGINT NOT NULL,
    
    currency VARCHAR(3) DEFAULT 'USD',
    
    -- Dates
    invoice_date TIMESTAMPTZ,
    due_date TIMESTAMPTZ,
    paid_at TIMESTAMPTZ,
    
    invoice_pdf_url TEXT,
    hosted_invoice_url TEXT,
    
    metadata JSONB DEFAULT '{}',
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_invoices_subscription_id ON invoices(subscription_id);
CREATE INDEX idx_invoices_team_id ON invoices(team_id);
CREATE INDEX idx_invoices_status ON invoices(status);
CREATE INDEX idx_invoices_invoice_date ON invoices(invoice_date DESC);
CREATE INDEX idx_invoices_stripe_invoice_id ON invoices(stripe_invoice_id);


-- ============================================================================
-- WEBHOOKS LOG
-- ============================================================================

CREATE TYPE webhook_source AS ENUM (
    'stripe',
    'twitter',
    'facebook',
    'instagram',
    'linkedin',
    'internal'
);

CREATE TABLE webhooks_log (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    source webhook_source NOT NULL,
    event_type VARCHAR(100) NOT NULL,
    
    payload JSONB NOT NULL,
    headers JSONB,
    
    processed BOOLEAN DEFAULT FALSE,
    processed_at TIMESTAMPTZ,
    
    response_status INT,
    response_body TEXT,
    error_message TEXT,
    
    idempotency_key VARCHAR(255) UNIQUE,
    
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_webhooks_log_source ON webhooks_log(source);
CREATE INDEX idx_webhooks_log_event_type ON webhooks_log(event_type);
CREATE INDEX idx_webhooks_log_processed ON webhooks_log(processed, created_at) WHERE processed = FALSE;
CREATE INDEX idx_webhooks_log_created_at ON webhooks_log(created_at DESC);


-- ============================================================================
-- JOB RUNS (Background jobs tracking)
-- ============================================================================

CREATE TYPE job_status AS ENUM (
    'pending',
    'running',
    'completed',
    'failed',
    'canceled'
);

CREATE TABLE job_runs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    job_name VARCHAR(100) NOT NULL,
    job_type VARCHAR(50) NOT NULL, -- 'scheduled_post', 'analytics_sync', 'token_refresh', etc.
    
    status job_status DEFAULT 'pending',
    
    -- Execution
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    duration_ms BIGINT, -- execution time in milliseconds
    
    worker_id VARCHAR(100),
    
    -- Context
    context JSONB DEFAULT '{}', -- job-specific params
    result JSONB, -- job output
    error_message TEXT,
    stack_trace TEXT,
    
    -- Retry
    attempt_number INT DEFAULT 1,
    max_attempts INT DEFAULT 3,
    next_retry_at TIMESTAMPTZ,
    
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_job_runs_job_name ON job_runs(job_name);
CREATE INDEX idx_job_runs_status ON job_runs(status);
CREATE INDEX idx_job_runs_created_at ON job_runs(created_at DESC);
CREATE INDEX idx_job_runs_next_retry_at ON job_runs(next_retry_at) WHERE status = 'failed' AND next_retry_at IS NOT NULL;


-- ============================================================================
-- TRIGGERS FOR UPDATED_AT
-- ============================================================================

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Apply trigger to all tables with updated_at
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_teams_updated_at BEFORE UPDATE ON teams FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_roles_updated_at BEFORE UPDATE ON roles FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_team_memberships_updated_at BEFORE UPDATE ON team_memberships FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_social_accounts_updated_at BEFORE UPDATE ON social_accounts FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_social_tokens_updated_at BEFORE UPDATE ON social_tokens FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_scheduled_posts_updated_at BEFORE UPDATE ON scheduled_posts FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_posts_updated_at BEFORE UPDATE ON posts FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_post_queue_updated_at BEFORE UPDATE ON post_queue FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_plans_updated_at BEFORE UPDATE ON plans FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_subscriptions_updated_at BEFORE UPDATE ON subscriptions FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_invoices_updated_at BEFORE UPDATE ON invoices FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_job_runs_updated_at BEFORE UPDATE ON job_runs FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();