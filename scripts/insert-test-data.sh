#!/bin/bash

# Test Data Bootstrap Script for Fern Platform
# ============================================
# 
# Purpose:
# This script populates the Fern Platform database with realistic test data for development,
# testing, and demonstration purposes. It creates a comprehensive dataset that simulates
# real-world test execution scenarios across multiple projects.
#
# What it does:
# 1. Connects to the PostgreSQL database running in Kubernetes using credentials from secrets
# 2. Creates 3 sample projects representing different types of applications:
#    - E-Commerce Frontend (React/TypeScript) - High pass rate (~87%)
#    - API Gateway Service (Go) - Lower pass rate (~51%) to simulate a struggling project
#    - Mobile Banking App (React Native) - High pass rate (~88%)
# 3. Generates test execution data for each project including:
#    - Multiple test runs with varying statuses (passed, failed, running)
#    - Test suites organized by type (unit, integration, e2e, etc.)
#    - Individual test specs with realistic names and error messages
#    - Historical data spanning several days for trend analysis
#
# Prerequisites:
# - kubectl configured with access to the Kubernetes cluster
# - Fern Platform deployed with PostgreSQL database
# - postgres-app secret containing database credentials
#
# Usage:
# ./scripts/insert-test-data.sh
#
# Note: This script is idempotent - running it multiple times will add more test data
# without removing existing data. Projects are created with ON CONFLICT handling.

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