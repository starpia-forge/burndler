# Burndler

A Docker Compose orchestration tool for merging, validating, and packaging multi-module applications for offline deployment.

## Overview

Burndler simplifies the deployment of complex Docker-based applications by:
- Merging multiple docker-compose files with intelligent namespace management
- Validating configurations against security and operational policies
- Creating self-contained offline installer packages with all required images

## Key Features

- **Compose Merging**: Combines multiple docker-compose.yml files with automatic namespace prefixing to prevent conflicts
- **Policy Validation**: Enforces security rules (no build directives, no privileged containers)
- **Offline Packaging**: Creates tar.gz installers with Docker images, compose files, and installation scripts
- **RBAC Support**: Role-based access control with Developer (read-write) and Engineer (read-only) permissions
- **Storage Flexibility**: Supports both S3-compatible and local filesystem storage

## Quick Start

### Prerequisites

- Go 1.24+
- Node.js 20+
- PostgreSQL 14+
- Docker and Docker Compose

### Development Setup

```bash
# Clone the repository
git clone https://github.com/burndler/burndler.git
cd burndler

# Copy environment configuration
cp .env.example .env

# Install dependencies
make deps-backend
make deps-frontend

# Start development environment
make dev
```

The application will be available at:
- Frontend: http://localhost:5173
- Backend API: http://localhost:8080

### Building

```bash
# Build all components
make build

# Build specific components
make build-backend
make build-frontend
```

### Testing

```bash
# Run all tests
make test

# Run specific test suites
make test-unit
make test-integration
```

## Architecture

Burndler follows a microservices architecture with specialized agents:

- **compose-merger**: Handles merging of multiple compose files
- **compose-linter**: Validates configurations against policies
- **image-packager**: Manages Docker image packaging
- **installer-packager**: Creates offline installer bundles
- **rbac-security**: Enforces authentication and authorization

## Configuration

Key environment variables:

```bash
# Database
DATABASE_URL=postgresql://user:pass@localhost/burndler

# Storage (S3 or Local)
STORAGE_TYPE=s3
S3_BUCKET=burndler-packages
S3_REGION=us-east-1

# JWT Authentication
JWT_SECRET=your-secret-key
```

## Deployment Policy

⚠️ **Important**: Burndler enforces a strict "prebuilt images only" policy:
- No `build:` directives in compose files
- No Dockerfiles in the repository
- All images must be pulled from registries

## API Documentation

API documentation is available in OpenAPI format at `/backend/openapi/openapi.yaml`

## Contributing

Please read our contributing guidelines before submitting pull requests.

## License

MIT License - see LICENSE file for details

---

*This project was built with [Claude Code](https://claude.ai/code) - Anthropic's AI-powered development assistant.*
