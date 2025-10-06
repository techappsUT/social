#!/bin/bash
# path: backend/scripts/fix-migration.sh

# Colors
BLUE='\033[0;34m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${BLUE}🔧 Migration Fix Script${NC}"
echo ""

# Check if Docker container is running
echo -e "${YELLOW}Checking if PostgreSQL container is running...${NC}"
if ! docker ps | grep -q socialqueue-postgres; then
    echo -e "${RED}❌ PostgreSQL container is not running.${NC}"
    echo "   Starting database services..."
    cd .. && docker-compose up -d postgres redis
    echo "   Waiting for PostgreSQL to be ready..."
    sleep 5
fi
echo -e "${GREEN}✓ PostgreSQL container is running${NC}"
echo ""

# Drop existing tables if they exist (in correct order)
echo -e "${YELLOW}Cleaning up existing tables...${NC}"
docker exec -i socialqueue-postgres psql -U socialqueue -d socialqueue_dev << EOF
-- Drop in reverse order of dependencies
DROP TABLE IF EXISTS refresh_tokens CASCADE;
DROP TABLE IF EXISTS users CASCADE;
DROP TABLE IF EXISTS teams CASCADE;
DROP FUNCTION IF EXISTS update_updated_at_column() CASCADE;
EOF

echo -e "${GREEN}✓ Cleanup complete${NC}"
echo ""

# Force migration version to 0 (clean state)
echo -e "${YELLOW}Resetting migration version...${NC}"
migrate -path ./migrations -database "postgres://socialqueue:socialqueue_dev_password@localhost:5432/socialqueue_dev?sslmode=disable" force 0 2>/dev/null
echo -e "${GREEN}✓ Migration version reset${NC}"
echo ""

# Run migrations
echo -e "${YELLOW}Running migrations...${NC}"
if make migrate-up; then
    echo ""
    echo -e "${GREEN}✅ Migration fixed successfully!${NC}"
    echo ""
    echo -e "${BLUE}Migration status:${NC}"
    make migrate-status
else
    echo ""
    echo -e "${RED}❌ Migration failed. Please check the error above.${NC}"
    exit 1
fi