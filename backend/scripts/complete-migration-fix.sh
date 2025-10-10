# path: backend/scripts/complete-migration-fix.sh

#!/bin/bash

BLUE='\033[0;34m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${BLUE}╔════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║  Complete Migration Fix & Reset       ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════╝${NC}"
echo ""

cd "$(dirname "$0")/.." || exit 1

# Step 1: Check PostgreSQL
echo -e "${YELLOW}Step 1: Checking PostgreSQL...${NC}"
if ! docker ps | grep -q socialqueue-postgres; then
    echo -e "${RED}   PostgreSQL not running. Starting...${NC}"
    cd .. && docker-compose up -d postgres redis && cd backend
    echo "   Waiting for PostgreSQL..."
    sleep 5
fi
echo -e "${GREEN}   ✓ PostgreSQL is running${NC}"
echo ""

# Step 2: Backup and consolidate migrations
echo -e "${YELLOW}Step 2: Consolidating migration files...${NC}"
cd migrations

# Backup
mkdir -p ../migrations_backup
cp *.sql ../migrations_backup/ 2>/dev/null || true
echo -e "   ${GREEN}✓${NC} Backup created"

# Remove duplicates
rm -f 001_create_users_table.sql 2>/dev/null
rm -f 20240101000001_initial_auth_schema.up.sql 2>/dev/null
rm -f 20240101000001_initial_auth_schema.down.sql 2>/dev/null
rm -f 000005_create_social_tokens.up.sql 2>/dev/null
rm -f 000005_create_social_tokens.down.sql 2>/dev/null
echo -e "   ${GREEN}✓${NC} Removed duplicate files"

# Rename to proper format
if [ -f "0001_initial_schema.sql" ]; then
    mv 0001_initial_schema.sql 00000000000001_initial_schema.up.sql
    echo -e "   ${GREEN}✓${NC} Renamed to proper format"
fi

cd ..
echo ""

# Step 3: Clean database
echo -e "${YELLOW}Step 3: Cleaning database...${NC}"
docker exec -i socialqueue-postgres psql -U socialqueue -d socialqueue_dev << 'EOF'
-- Drop schema_migrations to start fresh
DROP TABLE IF EXISTS schema_migrations CASCADE;

-- Drop all tables
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
DROP TABLE IF EXISTS roles CASCADE;
DROP TABLE IF EXISTS teams CASCADE;
DROP TABLE IF EXISTS users CASCADE;
DROP TABLE IF EXISTS refresh_tokens CASCADE;
DROP TABLE IF EXISTS dead_letter_queue CASCADE;

-- Drop functions
DROP FUNCTION IF EXISTS update_updated_at_column() CASCADE;

-- Drop enums
DROP TYPE IF EXISTS job_status CASCADE;
DROP TYPE IF EXISTS webhook_source CASCADE;
DROP TYPE IF EXISTS invoice_status CASCADE;
DROP TYPE IF EXISTS subscription_status CASCADE;
DROP TYPE IF EXISTS plan_interval CASCADE;
DROP TYPE IF EXISTS event_type CASCADE;
DROP TYPE IF EXISTS queue_status CASCADE;
DROP TYPE IF EXISTS attachment_type CASCADE;
DROP TYPE IF EXISTS post_status CASCADE;
DROP TYPE IF EXISTS social_account_status CASCADE;
DROP TYPE IF EXISTS social_platform CASCADE;
EOF

echo -e "${GREEN}   ✓ Database cleaned${NC}"
echo ""

# Step 4: Verify migration files
echo -e "${YELLOW}Step 4: Verifying migration files...${NC}"
cd migrations
echo -e "${BLUE}   Current migrations:${NC}"
for file in *.sql; do
    if [[ $file =~ ^[0-9]+_[a-z0-9_]+\.(up|down)\.sql$ ]]; then
        echo -e "   ${GREEN}✓${NC} $file"
    else
        echo -e "   ${RED}✗${NC} $file (invalid format)"
    fi
done
cd ..
echo ""

# Step 5: Run migrations
echo -e "${YELLOW}Step 5: Running migrations...${NC}"
if make migrate-up; then
    echo ""
    echo -e "${GREEN}╔════════════════════════════════════════╗${NC}"
    echo -e "${GREEN}║  ✅ Migration Success!                ║${NC}"
    echo -e "${GREEN}╚════════════════════════════════════════╝${NC}"
    echo ""
    echo -e "${BLUE}Migration Status:${NC}"
    make migrate-status
    echo ""
    echo -e "${BLUE}Next steps:${NC}"
    echo "  1. Verify tables: make db-shell"
    echo "  2. Run the API: make run"
    echo "  3. Seed data (optional): make seed"
else
    echo ""
    echo -e "${RED}╔════════════════════════════════════════╗${NC}"
    echo -e "${RED}║  ❌ Migration Failed                  ║${NC}"
    echo -e "${RED}╚════════════════════════════════════════╝${NC}"
    echo ""
    echo -e "${YELLOW}Troubleshooting:${NC}"
    echo "  1. Check migration files in ./migrations/"
    echo "  2. Ensure database is running: docker ps"
    echo "  3. Check logs: docker logs socialqueue-postgres"
    echo "  4. Manual fix: cd migrations && ls -la"
    exit 1
fi