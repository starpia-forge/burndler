# Burndler

[![CI](https://github.com/starpia-forge/burndler/actions/workflows/ci.yaml/badge.svg)](https://github.com/starpia-forge/burndler/actions/workflows/ci.yaml)
[![Release](https://github.com/starpia-forge/burndler/actions/workflows/release.yaml/badge.svg)](https://github.com/starpia-forge/burndler/actions/workflows/release.yaml)
[![GitHub release (latest SemVer)](https://img.shields.io/github/v/release/starpia-forge/burndler)](https://github.com/starpia-forge/burndler/releases/latest)
[![Docker Image](https://img.shields.io/badge/docker-ghcr.io%2Fstarpia--forge%2Fburndler-blue)](https://github.com/starpia-forge/burndler/pkgs/container/burndler)

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

## Release Management

### Development Workflow

Burndler uses automated semantic versioning based on [Conventional Commits](https://www.conventionalcommits.org/):

- **feat:** New feature (minor version bump)
- **fix:** Bug fix (patch version bump)
- **BREAKING CHANGE:** Major version bump
- **chore/docs/style/refactor/test:** No version bump

### Release Process

1. **Automatic Releases**: Triggered by pushes to `main` branch
2. **Manual Releases**: Use `make release-local` for local testing
3. **Version Check**: Run `make version` to see current version info
4. **Pre-commit Hooks**: Run `make pre-commit-install` to set up quality gates

### Available Artifacts

- **Binaries**: Linux AMD64 binaries attached to GitHub releases
- **Docker Images**: `ghcr.io/starpia-forge/burndler:latest` and tagged versions
- **Offline Installers**: Complete deployment packages in release assets

### Development Commands

Essential commands for contributors:

```bash
# Setup development environment
make setup && make pre-commit-install

# Run tests and quality checks
make test && make lint

# Build and verify
make build && make version

# Check release readiness
make release-check
```

## Contributing

Please read our contributing guidelines before submitting pull requests. Ensure all commits follow the [Conventional Commits](https://www.conventionalcommits.org/) specification.

## License

MIT License - see LICENSE file for details

---

*This project was built with [Claude Code](https://claude.ai/code) - Anthropic's AI-powered development assistant.*
