#!/bin/bash

# Quick script to create indexes for customer_profile table
# This addresses the query timeout issue

set -e

# Load environment variables
if [ -f .env ]; then
    export $(cat .env | grep -v '^#' | xargs)
fi

DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_USER="${DB_USER:-customerdb_user}"
DB_NAME="${DB_NAME:-CustomerDB}"
DB_PASSWORD="${DB_PWD:-$DB_PASSWORD}"

echo "Creating indexes on customer_profile table..."
echo "This will fix the query timeout issue."
echo ""

PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -f internal/database/migrations/004_create_customer_profile_indexes.sql

echo ""
echo "✓ Indexes created successfully!"
echo "The FindDuplicateIndividual query should now run much faster."

