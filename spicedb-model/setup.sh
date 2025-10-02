#!/bin/bash
set -e

echo "🚀 Setting up SpiceDB with schema and data..."

# Start SpiceDB
cd "$(dirname "$0")/spicedb-config"
echo "📦 Starting SpiceDB container..."
docker-compose up -d
cd ..

# Install zed CLI if not present
if ! command -v zed &> /dev/null; then
  echo "📥 Installing zed CLI via Homebrew..."
  brew install authzed/tap/zed
else
  echo "✅ zed CLI already installed."
fi

# Wait for SpiceDB to be ready
echo "⏳ Waiting for SpiceDB to be ready on localhost:50051..."
for i in {1..30}; do
  if nc -z localhost 50051 2>/dev/null; then
    echo "✅ SpiceDB is up and running."
    break
  fi
  if [ $i -eq 30 ]; then
    echo "❌ ERROR: SpiceDB did not start on port 50051 within 30 seconds." >&2
    echo "   Check if Docker is running and port 50051 is available." >&2
    exit 1
  fi
  sleep 1
done

# Load schema
echo "📋 Loading schema from model.zaml..."
zed schema write model.zaml
echo "✅ Schema loaded successfully."

# Load relationships using the robust loader
echo "📊 Loading relationships..."
chmod +x load_relationships.sh
./load_relationships.sh

echo ""
echo "🎉 Setup complete! SpiceDB is ready with schema and data."
echo ""
echo "Next steps:"
echo "  • Run tests: bash run_permission_checks.sh"
echo "  • Check SpiceDB: zed schema read"
echo "  • View relationships: zed relationship read" 