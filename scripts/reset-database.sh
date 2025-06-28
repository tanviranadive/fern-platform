#!/bin/bash

# Script to reset the database for fresh deployment

set -e

echo "🔄 Resetting fern-platform database..."

# Get the postgres pod
POSTGRES_POD=$(kubectl -n fern-platform get pods -l app=postgres -o jsonpath='{.items[0].metadata.name}')

if [ -z "$POSTGRES_POD" ]; then
    echo "❌ No postgres pod found"
    exit 1
fi

echo "📊 Using postgres pod: $POSTGRES_POD"

# Drop and recreate the database
echo "🗑️  Dropping existing database..."
kubectl -n fern-platform exec -i "$POSTGRES_POD" -- psql -U postgres -c "DROP DATABASE IF EXISTS app;"

echo "✨ Creating fresh database..."
kubectl -n fern-platform exec -i "$POSTGRES_POD" -- psql -U postgres -c "CREATE DATABASE app OWNER app;"

echo "🔑 Granting permissions..."
kubectl -n fern-platform exec -i "$POSTGRES_POD" -- psql -U postgres -c "GRANT ALL PRIVILEGES ON DATABASE app TO app;"

echo "✅ Database reset complete!"
echo ""
echo "🚀 Now restart the fern-platform pod:"
echo "   kubectl -n fern-platform rollout restart deployment fern-platform"