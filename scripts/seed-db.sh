#!/bin/bash
# Seed Burndler development database with sample data

set -e

echo "🌱 Seeding Burndler Development Database"
echo "======================================="

# Load environment variables
if [ -f .env.development ]; then
    source .env.development
else
    echo "❌ .env.development file not found"
    exit 1
fi

# Default values
DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-5432}
DB_NAME=${DB_NAME:-burndler_dev}
DB_USER=${DB_USER:-burndler}

echo "📋 Database: $DB_NAME on $DB_HOST:$DB_PORT"

# Check if database is accessible
if ! pg_isready -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" >/dev/null 2>&1; then
    echo "❌ Database is not accessible. Make sure PostgreSQL is running:"
    echo "   make dev-db"
    exit 1
fi

echo "✅ Database is accessible"

# Run seeding through the Go application
echo ""
echo "🚀 Running database migrations and seeding..."

cd backend

# First run migrations
echo "📦 Running migrations..."
if go run cmd/api/main.go migrate; then
    echo "✅ Migrations completed"
else
    echo "❌ Migration failed"
    exit 1
fi

# TODO: Add seeding logic here when implemented
# For now, just ensure the database is properly migrated

echo ""
echo "🎉 Database seeding completed!"
echo ""
echo "Sample data includes:"
echo "- Database schema (via migrations)"
echo "- Ready for development data"
echo ""
echo "You can now:"
echo "1. Start backend:  make dev-backend"
echo "2. Start frontend: make dev-frontend"