-- path: backend/migrations/20240101000001_initial_schema.down.sql

-- Drop triggers
DROP TRIGGER IF EXISTS update_job_runs_updated_at ON job_runs;
DROP TRIGGER IF EXISTS update_invoices_updated_at ON invoices;
DROP TRIGGER IF EXISTS update_subscriptions_updated_at ON subscriptions;
DROP TRIGGER IF EXISTS update_plans_updated_at ON plans;
DROP TRIGGER IF EXISTS update_post_queue_updated_at ON post_queue;
DROP TRIGGER IF EXISTS update_posts_updated_at ON posts;
DROP TRIGGER IF EXISTS update_scheduled_posts_updated_at ON scheduled_posts;
DROP TRIGGER IF EXISTS update_social_tokens_updated_at ON social_tokens;
DROP TRIGGER IF EXISTS update_social_accounts_updated_at ON social_accounts;
DROP TRIGGER IF EXISTS update_team_memberships_updated_at ON team_memberships;
DROP TRIGGER IF EXISTS update_roles_updated_at ON roles;
DROP TRIGGER IF EXISTS update_teams_updated_at ON teams;
DROP TRIGGER IF EXISTS update_users_updated_at ON users;

-- Drop tables in reverse dependency order
DROP TABLE IF EXISTS job_runs CASCADE;
DROP TABLE IF EXISTS webhooks_log CASCADE;
DROP TABLE IF EXISTS invoices CASCADE;
DROP TABLE IF EXISTS subscriptions CASCADE;
DROP TABLE IF EXISTS plans CASCADE;
DROP TABLE IF EXISTS analytics_events CASCADE;
DROP TABLE IF EXISTS post_queue CASCADE;
DROP TABLE IF EXISTS posts CASCADE;
DROP TABLE IF EXISTS post_attachments CASCADE;
DROP TABLE IF EXISTS scheduled_posts CASCADE;
DROP TABLE IF EXISTS social_tokens CASCADE;
DROP TABLE IF EXISTS social_accounts CASCADE;
DROP TABLE IF EXISTS team_memberships CASCADE;
DROP TABLE IF EXISTS refresh_tokens CASCADE;
DROP TABLE IF EXISTS roles CASCADE;
DROP TABLE IF EXISTS teams CASCADE;
DROP TABLE IF EXISTS users CASCADE;

-- Drop functions
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop enums
DROP TYPE IF EXISTS job_status;
DROP TYPE IF EXISTS webhook_source;
DROP TYPE IF EXISTS invoice_status;
DROP TYPE IF EXISTS subscription_status;
DROP TYPE IF EXISTS plan_interval;
DROP TYPE IF EXISTS event_type;
DROP TYPE IF EXISTS queue_status;
DROP TYPE IF EXISTS attachment_type;
DROP TYPE IF EXISTS post_status;
DROP TYPE IF EXISTS social_account_status;
DROP TYPE IF EXISTS social_platform;

-- Drop extensions
DROP EXTENSION IF EXISTS pg_trgm;
DROP EXTENSION IF EXISTS "uuid-ossp";