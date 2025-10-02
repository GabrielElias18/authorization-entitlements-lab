#!/bin/bash
set -e

echo "🔄 Resetting GraphQL data..."

# Database connection settings
PGHOST=localhost
PGPORT=5434
PGUSER=postgres
PGDATABASE=entitlements
PGPASSWORD=postgres
export PGPASSWORD

cd "$(dirname "$0")"

# Check if PostgreSQL is running
if ! psql -h $PGHOST -p $PGPORT -U $PGUSER -d $PGDATABASE -c '\q' 2>/dev/null; then
  echo "❌ PostgreSQL is not running on port 5434."
  echo "   Please start PostgreSQL first: bash setup.sh"
  exit 1
fi

# Clear all data
echo "🗑️  Clearing all existing data..."
psql -h $PGHOST -p $PGPORT -U $PGUSER -d $PGDATABASE -c "TRUNCATE \"POA\", \"Payment\", \"Statement\", \"Account\", \"User\", \"Role\", \"Organization\" RESTART IDENTITY CASCADE;"

# Reload canonical data
echo "📊 Reloading canonical data..."
psql -h $PGHOST -p $PGPORT -U $PGUSER -d $PGDATABASE -f init-scripts/01-seed-data.sql

echo ""
echo "🎉 Reset complete! GraphQL has fresh data."
echo ""
echo "Next steps:"
echo "  • Run tests: bash run_permission_checks.sh"
echo "  • Check data: psql -h localhost -p 5434 -U postgres -d entitlements" 