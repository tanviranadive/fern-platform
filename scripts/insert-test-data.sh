#!/bin/bash

# Script to insert test data into fern-platform database
# This script uses kubectl to get database credentials from secrets and then applies test data

set -e

echo "🔍 Getting database connection details from Kubernetes secrets..."

# Get database credentials from secret
DB_HOST=$(kubectl get secret postgres-app -n fern-platform -o jsonpath='{.data.host}' | base64 -d)
DB_PORT=$(kubectl get secret postgres-app -n fern-platform -o jsonpath='{.data.port}' | base64 -d)
DB_NAME=$(kubectl get secret postgres-app -n fern-platform -o jsonpath='{.data.dbname}' | base64 -d)
DB_USER=$(kubectl get secret postgres-app -n fern-platform -o jsonpath='{.data.username}' | base64 -d)
DB_PASSWORD=$(kubectl get secret postgres-app -n fern-platform -o jsonpath='{.data.password}' | base64 -d)

echo "📊 Database connection details retrieved"
echo "   Host: $DB_HOST"
echo "   Port: $DB_PORT"
echo "   Database: $DB_NAME"
echo "   User: $DB_USER"

# Create SQL file path
SQL_FILE="$(dirname "$0")/test-data.sql"

# Check if SQL file exists
if [ ! -f "$SQL_FILE" ]; then
    echo "❌ SQL file not found: $SQL_FILE"
    exit 1
fi

echo "📝 Applying test data from $SQL_FILE..."

# Find postgres pod
POSTGRES_POD=$(kubectl get pods -n fern-platform -l cnpg.io/cluster=postgres -o jsonpath='{.items[0].metadata.name}')

if [ -z "$POSTGRES_POD" ]; then
    echo "❌ No postgres pod found"
    exit 1
fi

echo "🚀 Using postgres pod: $POSTGRES_POD"

# Execute SQL file using stdin with password
kubectl exec -i -n fern-platform "$POSTGRES_POD" -- bash -c "PGPASSWORD='$DB_PASSWORD' psql -h '$DB_HOST' -p '$DB_PORT' -U '$DB_USER' -d '$DB_NAME'" < "$SQL_FILE"

echo "✅ Test data inserted successfully!"