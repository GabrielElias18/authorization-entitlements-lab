#!/bin/bash
set -e

echo "🚀 Setting up GraphQL with PostgreSQL and Prisma..."

# Database connection settings
PGHOST=localhost
PGPORT=5434
PGUSER=postgres
PGDATABASE=entitlements
PGPASSWORD=postgres
export PGPASSWORD

cd "$(dirname "$0")"

# Stop any running containers
echo "📦 Stopping any running containers..."
docker-compose down

# Start PostgreSQL
echo "🐘 Starting PostgreSQL..."
docker-compose up -d postgres

# Wait for PostgreSQL to be ready
echo "⏳ Waiting for PostgreSQL to be ready on localhost:5434..."
for i in {1..30}; do
  if psql -h $PGHOST -p $PGPORT -U $PGUSER -d $PGDATABASE -c '\q' 2>/dev/null; then
    echo "✅ PostgreSQL is up and running."
    break
  fi
  if [ $i -eq 30 ]; then
    echo "❌ ERROR: PostgreSQL did not start on port 5434 within 30 seconds." >&2
    echo "   Check if Docker is running and port 5434 is available." >&2
    exit 1
  fi
  sleep 1
done

# Install dependencies
echo "📥 Installing Node.js dependencies..."
npm install

# Run Prisma migrations
echo "🔄 Running Prisma migrations..."
npx prisma migrate deploy

# Load canonical data
echo "📊 Loading canonical data..."
psql -h $PGHOST -p $PGPORT -U $PGUSER -d $PGDATABASE -f init-scripts/01-seed-data.sql

echo ""
echo "🎉 Setup complete! GraphQL server is ready."
echo ""
echo "Next steps:"
echo "  • Start server: npm run dev"
echo "  • Run tests: bash run_permission_checks.sh"
echo "  • Check data: psql -h localhost -p 5434 -U postgres -d entitlements" 