# path: backend/scripts/consolidate-migrations.sh

#!/bin/bash

BLUE='\033[0;34m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${BLUE}🧹 Consolidating Migration Files${NC}"
echo ""

cd "$(dirname "$0")/../migrations" || exit 1

# Backup existing migrations
echo -e "${YELLOW}1. Creating backup...${NC}"
mkdir -p ../migrations_backup
cp *.sql ../migrations_backup/ 2>/dev/null || true
echo -e "${GREEN}   ✓ Backup created in migrations_backup/${NC}"
echo ""

# Remove all duplicate/old migration files
echo -e "${YELLOW}2. Removing duplicate migration files...${NC}"
rm -f 001_create_users_table.sql
rm -f 20240101000001_initial_auth_schema.up.sql
rm -f 20240101000001_initial_auth_schema.down.sql
rm -f 000005_create_social_tokens.up.sql
rm -f 000005_create_social_tokens.down.sql
echo -e "${GREEN}   ✓ Duplicates removed${NC}"
echo ""

# Rename 0001_initial_schema to proper format
echo -e "${YELLOW}3. Renaming to proper format...${NC}"
if [ -f "0001_initial_schema.sql" ]; then
    mv 0001_initial_schema.sql 00000000000001_initial_schema.up.sql
    echo "   Renamed: 0001_initial_schema.sql → 00000000000001_initial_schema.up.sql"
fi
echo -e "${GREEN}   ✓ Renamed to proper format${NC}"
echo ""

echo -e "${BLUE}Current migration files:${NC}"
ls -1 *.sql 2>/dev/null | sort
echo ""

cd ..