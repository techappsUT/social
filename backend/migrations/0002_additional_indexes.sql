-- path: backend/migrations/0002_additional_indexes.sql

-- Composite indexes for common queries

-- Dashboard: team's recent posts
CREATE INDEX idx_scheduled_posts_team_created ON scheduled_posts(team_id, created_at DESC) WHERE deleted_at IS NULL;

-- Analytics: team's posts within date range
CREATE INDEX idx_posts_team_published ON posts(team_id, published_at DESC);

-- Queue processing: fetch next jobs
CREATE INDEX idx_post_queue_next_jobs ON post_queue(scheduled_for, priority) 
    WHERE status = 'pending' AND lock_expires_at IS NULL;

-- Social account lookup by team and platform
CREATE INDEX idx_social_accounts_team_platform ON social_accounts(team_id, platform) WHERE deleted_at IS NULL;

-- User's teams
CREATE INDEX idx_team_memberships_user_team ON team_memberships(user_id, team_id) WHERE deleted_at IS NULL;

-- Analytics aggregation by team and date (FIXED - removed cast)
CREATE INDEX idx_analytics_events_team_date ON analytics_events(team_id, event_timestamp);

-- Subscription expiring soon
CREATE INDEX idx_subscriptions_expiring ON subscriptions(current_period_end) 
    WHERE status = 'active' AND cancel_at_period_end = FALSE;

-- Full-text search on posts (optional, for search feature)
CREATE INDEX idx_scheduled_posts_content_trgm ON scheduled_posts USING gin(content gin_trgm_ops);