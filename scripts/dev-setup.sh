#!/bin/bash
# Burndler Development Environment Setup Script

set -e

echo "🚀 Burndler Development Environment Setup"
echo "========================================"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to print colored output
print_status() {
    echo -e "${GREEN}✅${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}⚠️${NC} $1"
}

print_error() {
    echo -e "${RED}❌${NC} $1"
}

print_info() {
    echo -e "${BLUE}ℹ️${NC} $1"
}

# Check prerequisites
echo ""
echo "🔍 Checking prerequisites..."

# Check Go
if command_exists go; then
    GO_VERSION=$(go version | cut -d' ' -f3)
    print_status "Go installed: $GO_VERSION"
else
    print_error "Go is not installed. Please install Go 1.22 or later."
    exit 1
fi

# Check Node.js
if command_exists node; then
    NODE_VERSION=$(node --version)
    print_status "Node.js installed: $NODE_VERSION"
else
    print_error "Node.js is not installed. Please install Node.js 20 or later."
    exit 1
fi

# Check npm
if command_exists npm; then
    NPM_VERSION=$(npm --version)
    print_status "npm installed: v$NPM_VERSION"
else
    print_error "npm is not installed."
    exit 1
fi

# Check Docker
if command_exists docker; then
    print_status "Docker is available"
else
    print_error "Docker is not installed. Please install Docker for PostgreSQL."
    exit 1
fi

# Check Docker Compose
if command_exists docker-compose; then
    print_status "Docker Compose is available"
elif docker compose version >/dev/null 2>&1; then
    print_status "Docker Compose (v2) is available"
else
    print_error "Docker Compose is not installed."
    exit 1
fi

# Setup environment files
echo ""
echo "📝 Setting up environment files..."

if [ ! -f .env.development ]; then
    if [ -f .env.example ]; then
        cp .env.example .env.development
        print_status "Created .env.development from .env.example"
    else
        print_warning ".env.example not found, .env.development may need manual setup"
    fi
else
    print_status ".env.development already exists"
fi

# Setup frontend environment
if [ ! -f frontend/.env.development ]; then
    print_status "Frontend .env.development created"
else
    print_status "Frontend .env.development already exists"
fi

# Install Air for Go hot reload
echo ""
echo "🔥 Setting up Air for Go hot reload..."

if command_exists air; then
    print_status "Air is already installed"
else
    print_info "Installing Air..."
    if go install github.com/cosmtrek/air@latest; then
        print_status "Air installed successfully"
    else
        print_error "Failed to install Air"
        exit 1
    fi
fi

# Install backend dependencies
echo ""
echo "📦 Installing backend dependencies..."
cd backend
if go mod download; then
    print_status "Backend dependencies installed"
else
    print_error "Failed to install backend dependencies"
    exit 1
fi
cd ..

# Install frontend dependencies
echo ""
echo "📦 Installing frontend dependencies..."
cd frontend
if npm install; then
    print_status "Frontend dependencies installed"
else
    print_error "Failed to install frontend dependencies"
    exit 1
fi
cd ..

# Create tmp directory for Air
echo ""
echo "📁 Creating directories..."
mkdir -p tmp
print_status "Created tmp directory for Air"

# Make scripts executable
echo ""
echo "🔧 Setting up scripts..."
chmod +x scripts/*.sh
print_status "Made scripts executable"

# Setup complete
echo ""
echo "🎉 Development environment setup complete!"
echo ""
echo "Next steps:"
echo "1. Start PostgreSQL:  make dev-db"
echo "2. Start backend:     make dev-backend"
echo "3. Start frontend:    make dev-frontend"
echo ""
echo "Or run 'make dev' for guided setup."