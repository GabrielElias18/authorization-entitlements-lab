#!/bin/bash
set -e

echo "🔄 Resetting SpiceDB data..."

# Clear all existing relationships
for type in user org role account poa; do
  echo "🗑️  Clearing all $type relationships..."
  zed relationship bulk-delete "$type" --force
  sleep 1
  done

# Reload schema
echo "📋 Reloading schema..."
zed schema write model.zaml

# Load relationships using the robust loader
echo "📊 Reloading relationships..."
chmod +x load_relationships.sh
./load_relationships.sh

echo ""
echo "🎉 Reset complete! SpiceDB has fresh schema and data."
echo ""
echo "Next steps:"
echo "  • Run tests: bash run_permission_checks.sh"
echo "  • Check data: zed relationship read" 