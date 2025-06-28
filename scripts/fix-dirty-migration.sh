#!/bin/bash

# Script to fix dirty database migration state

set -e

echo "🔧 Fixing dirty database migration..."

# Get the postgres pod
POSTGRES_POD=$(kubectl -n fern-platform get pods -l app=postgres -o jsonpath='{.items[0].metadata.name}')

if [ -z "$POSTGRES_POD" ]; then
    echo "❌ No postgres pod found"
    exit 1
fi

echo "📊 Using postgres pod: $POSTGRES_POD"

# Connect to database and fix the dirty state
echo "🔄 Forcing migration version to 14..."

kubectl -n fern-platform exec -i "$POSTGRES_POD" -- psql -U fern -d fern_platform -c "UPDATE schema_migrations SET dirty = false WHERE version = 14;"

echo "✅ Migration state fixed!"
echo ""
echo "🚀 You can now restart the fern-platform pod:"
echo "   kubectl -n fern-platform delete pod -l app=fern-platform"