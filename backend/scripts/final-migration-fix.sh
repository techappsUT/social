# path: backend/scripts/final-migration-fix.sh

#!/bin/bash

BLUE='\033[0;34m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${BLUE}🔧 Final Migration Fix${NC}"
echo ""

cd "$(dirname "$0")/../migrations" || exit 1

# Show current state
echo -e "${YELLOW}Current migration files:${NC}"
ls -la *.sql
echo ""

# Remove ALL old migration files
echo -e "${YELLOW}Removing all migration files...${NC}"
rm -f *.sql
echo -e "${GREEN}   ✓ Cleared${NC}"
echo ""

# Check if backup exists
if [ -d "../migrations_backup" ]; then
    echo -e "${YELLOW}Restoring from backup...${NC}"
    # Copy the complete initial schema (which was 0001_initial_schema.sql originally)
    if [ -f "../migrations_backup/0001_initial_schema.sql" ]; then
        cp ../migrations_backup/0001_initial_schema.sql ./20240101000001_initial_schema.up.sql
        echo -e "${GREEN}   ✓ Restored initial schema as 20240101000001_initial_schema.up.sql${NC}"
    fi
    
    # Copy the down migration if it exists
    if [ -f "../migrations_backup/0001_initial_schema.down.sql" ]; then
        cp ../migrations_backup/0001_initial_schema.down.sql ./20240101000001_initial_schema.down.sql
        echo -e "${GREEN}   ✓ Restored down migration${NC}"
    fi
fi

echo ""
echo -e "${BLUE}Final migration files:${NC}"
ls -la *.sql
echo ""

cd ..

# Clean database
echo -e "${YELLOW}Cleaning database...${NC}"
docker exec -i socialqueue-postgres psql -U socialqueue -d socialqueue_dev << 'EOF'
DROP SCHEMA public CASCADE;
CREATE SCHEMA public;
GRANT ALL ON SCHEMA public TO socialqueue;
GRANT ALL ON SCHEMA public TO public;
EOF

echo -e "${GREEN}   ✓ Database cleaned${NC}"
echo ""

# Run migrations
echo -e "${YELLOW}Running migrations...${NC}"
if make migrate-up; then
    echo ""
    echo -e "${GREEN}✅ Success!${NC}"
    make migrate-status
else
    echo ""
    echo -e "${RED}❌ Failed${NC}"
    exit 1
fi