#!/bin/bash
# Reset Burndler development environment

set -e

echo "🔄 Resetting Burndler Development Environment"
echo "============================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_status() {
    echo -e "${GREEN}✅${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}⚠️${NC} $1"
}

print_info() {
    echo -e "${BLUE}ℹ️${NC} $1"
}

# Stop any running services
echo ""
echo "🛑 Stopping development services..."

# Stop PostgreSQL if running
if docker ps | grep burndler_postgres_dev > /dev/null; then
    print_info "Stopping PostgreSQL container..."
    docker-compose -f compose/postgres.compose.yaml --env-file .env.development down -v
    print_status "PostgreSQL stopped and volumes removed"
else
    print_info "PostgreSQL container not running"
fi

# Clean build artifacts
echo ""
echo "🧹 Cleaning build artifacts..."

# Clean backend artifacts
if [ -d "tmp" ]; then
    rm -rf tmp/*
    print_status "Cleaned tmp directory"
fi

if [ -d "backend/coverage.out" ]; then
    rm -f backend/coverage.*
    print_status "Cleaned backend coverage files"
fi

if [ -f "build-errors.log" ]; then
    rm -f build-errors.log
    print_status "Cleaned Air build error logs"
fi

# Clean frontend artifacts
if [ -d "frontend/dist" ]; then
    rm -rf frontend/dist
    print_status "Cleaned frontend dist directory"
fi

if [ -d "frontend/coverage" ]; then
    rm -rf frontend/coverage
    print_status "Cleaned frontend coverage directory"
fi

if [ -d "frontend/node_modules/.vite" ]; then
    rm -rf frontend/node_modules/.vite
    print_status "Cleaned Vite cache"
fi

# Clean dist directory
if [ -d "dist" ]; then
    rm -rf dist/*
    print_status "Cleaned project dist directory"
fi

# Clean logs
find . -name "*.log" -type f -delete 2>/dev/null || true
print_status "Cleaned log files"

# Restart fresh environment
echo ""
echo "🚀 Starting fresh environment..."

# Start PostgreSQL
print_info "Starting PostgreSQL with fresh data..."
make dev-db

# Wait for database to be ready
sleep 3

print_status "Environment reset complete!"
echo ""
echo "Your development environment is now clean and ready."
echo ""
echo "Next steps:"
echo "1. Start backend:  make dev-backend"
echo "2. Start frontend: make dev-frontend"