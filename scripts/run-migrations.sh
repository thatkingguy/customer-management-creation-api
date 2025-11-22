#!/bin/bash

# Script to run database migrations
# Usage: ./scripts/run-migrations.sh [migration_file]

set -e

# Load environment variables
if [ -f .env ]; then
    export $(cat .env | grep -v '^#' | xargs)
fi

# Database connection details
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_USER="${DB_USER:-customerdb_user}"
DB_NAME="${DB_NAME:-CustomerDB}"
DB_PASSWORD="${DB_PWD:-$DB_PASSWORD}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}Running database migrations...${NC}"
echo "Host: $DB_HOST"
echo "Port: $DB_PORT"
echo "Database: $DB_NAME"
echo "User: $DB_USER"
echo ""

# If specific migration file provided, run only that
if [ -n "$1" ]; then
    MIGRATION_FILE="$1"
    if [ ! -f "$MIGRATION_FILE" ]; then
        echo -e "${RED}Error: Migration file not found: $MIGRATION_FILE${NC}"
        exit 1
    fi
    echo -e "${YELLOW}Running migration: $MIGRATION_FILE${NC}"
    PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -f "$MIGRATION_FILE"
    echo -e "${GREEN}Migration completed successfully!${NC}"
    exit 0
fi

# Run all migrations in order
MIGRATION_DIR="internal/database/migrations"
MIGRATIONS=(
    "001_create_sequences.sql"
    "002_create_indexes.sql"
    "003_create_customer_profile_lookup_indexes.sql"
    "004_create_customer_profile_indexes.sql"
    "005_create_sector_industry_target_indexes.sql"
    "006_create_unique_constraints_for_lookup.sql"
    "007_create_request_and_interim_indexes.sql"
)

for migration in "${MIGRATIONS[@]}"; do
    MIGRATION_FILE="$MIGRATION_DIR/$migration"
    if [ -f "$MIGRATION_FILE" ]; then
        echo -e "${YELLOW}Running migration: $migration${NC}"
        PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -f "$MIGRATION_FILE"
        if [ $? -eq 0 ]; then
            echo -e "${GREEN}✓ $migration completed${NC}"
        else
            echo -e "${RED}✗ $migration failed${NC}"
            exit 1
        fi
    else
        echo -e "${YELLOW}Warning: Migration file not found: $MIGRATION_FILE${NC}"
    fi
done

echo ""
echo -e "${GREEN}All migrations completed successfully!${NC}"



