# ADR-004: Repository Workflow

## Status
Accepted

## Context
Define development workflow, build automation, and testing strategy for the Burndler mono-repo.

## Decision

### Make Targets

```makefile
# Development
make dev          # Start local development environment
make dev-backend  # Run backend only (Go + PostgreSQL)
make dev-frontend # Run frontend only (React dev server)

# Build
make build        # Build all components
make build-backend # Build Go binary
make build-frontend # Build React production bundle
make build-tools  # Build CLI tools

# Testing
make test         # Run all tests
make test-unit    # Unit tests only
make test-integration # Integration tests
make test-e2e     # End-to-end tests with Playwright

# Quality
make lint         # Run all linters
make lint-go      # golangci-lint
make lint-js      # ESLint + Prettier
make lint-compose # Validate compose files (no build:, etc.)

# Operations
make merge        # Test compose merge functionality
make package      # Create offline installer
make clean        # Remove build artifacts
```

### Development Compose Policy

#### docker-compose.dev.yml
```yaml
# Development environment rules:
# 1. NO build: directives - use prebuilt images only
# 2. Use image@sha256 for reproducibility
# 3. Mount source code as volumes for hot reload
# 4. Expose all ports for debugging
# 5. Use .env for configuration

services:
  postgres:
    image: postgres:15@sha256:...
    environment:
      POSTGRES_DB: burndler_dev
    volumes:
      - postgres_data:/var/lib/postgresql/data
    ports:
      - "5432:5432"

  backend:
    image: golang:1.22@sha256:...
    working_dir: /app
    volumes:
      - ./backend:/app
      - go_modules:/go/pkg/mod
    command: go run cmd/api/main.go
    env_file: .env
    ports:
      - "8080:8080"

  frontend:
    image: node:20@sha256:...
    working_dir: /app
    volumes:
      - ./frontend:/app
      - node_modules:/app/node_modules
    command: npm run dev
    ports:
      - "3000:3000"

volumes:
  postgres_data:
  go_modules:
  node_modules:
```

### Testing Strategy

#### Priority Order
1. **Compose Operations** (Critical Path)
   - Merge logic with namespace prefixing
   - Variable substitution cascading
   - Conflict detection and resolution

2. **Lint Policy** (Quality Gate)
   - Forbidden directive detection (build:)
   - Environment variable validation
   - Port collision detection
   - Reference integrity checking

3. **Packaging Pipeline** (Delivery)
   - Offline installer generation
   - Image digest resolution
   - Manifest generation
   - Installation script validation

4. **RBAC Implementation** (Security)
   - JWT middleware functionality
   - Role-based route protection
   - Permission enforcement

#### Test Organization
```
test/
├── unit/
│   ├── backend/      # Go unit tests
│   ├── frontend/     # React component tests
│   └── tools/        # CLI tool tests
│
├── integration/
│   ├── compose/      # Multi-module merge scenarios
│   ├── storage/      # S3/Local FS switching
│   └── rbac/         # Auth flow tests
│
├── e2e/
│   ├── workflows/    # Complete user journeys
│   └── offline/      # Offline installer validation
│
└── fixtures/
    ├── compose/      # Sample compose files
    ├── packages/     # Test artifacts
    └── config/       # Test configurations
```

### CI/CD Pipeline

```yaml
# .github/workflows/ci.yml or .gitlab-ci.yml concept

stages:
  - validate   # Lint, format check
  - test      # Unit + integration tests
  - build     # Create binaries and bundles
  - package   # Generate offline installer
  - e2e       # Full workflow validation

validate:
  - make lint
  - Verify no build: directives
  - Check for committed secrets

test:
  - make test-unit
  - make test-integration
  - Coverage report

build:
  - make build
  - Store artifacts

package:
  - make package
  - Validate manifest.json
  - Test install.sh in container

e2e:
  - Deploy test environment
  - Run Playwright tests
  - Offline installation test
```

### Development Guidelines

1. **Branch Strategy**
   - `main` - stable, deployable
   - `develop` - integration branch
   - `feature/*` - new features
   - `fix/*` - bug fixes

2. **Commit Standards**
   - Conventional commits format
   - Sign commits for production changes
   - Reference issues/ADRs in messages

3. **Code Review Requirements**
   - All changes via pull request
   - Automated checks must pass
   - Security review for auth/RBAC changes

4. **Local Development Setup**
   ```bash
   # Initial setup
   cp .env.example .env
   make dev

   # Run tests before commit
   make lint test

   # Test packaging locally
   make package
   ```

## Consequences

### Positive
- Reproducible builds with pinned images
- Clear testing priorities aligned with business value
- Consistent development environment across team
- Automated quality gates prevent policy violations

### Negative
- No dynamic image building may slow iteration
- Requires discipline to maintain image digests
- Additional overhead for image management

## References
- ADR-001: Compose merge strategy
- ADR-002: Lint policy
- ADR-003: Packaging format