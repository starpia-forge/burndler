# Burndler Repository Structure

## Directory Layout

```
burndler/
├── backend/                 # Go API server (Gin + GORM)
│   ├── cmd/                # Application entrypoints
│   ├── internal/           # Private application code
│   ├── pkg/                # Public packages
│   ├── openapi/            # OpenAPI specifications
│   └── docs/               # Backend documentation
│
├── frontend/               # React TypeScript application
│   ├── src/               # Source code (components, hooks, utils)
│   ├── public/            # Static assets
│   └── tests/             # Frontend tests
│
├── ops/                    # Operational tooling
│   ├── adr/               # Architecture Decision Records (ADR-001..004)
│   ├── compose/           # Docker Compose templates for merging
│   └── scripts/           # Build, lint, and packaging scripts
│
├── tools/                  # CLI tools for compose operations
│   ├── merge/             # Compose merger implementation
│   ├── lint/              # Compose linter implementation
│   └── package/           # Offline installer packager
│
├── test/                   # Integration and E2E tests
│   ├── integration/       # Multi-component tests
│   └── fixtures/          # Test data and compose samples
│
├── dist/                   # Build artifacts (gitignored)
│   ├── images/            # Docker image .tar files
│   └── installers/        # Packaged offline installers
│
└── repo/                   # Repository documentation
    └── FILES.md           # This file

## Key Files

- `.env.example`           # Environment template (copy to .env)
- `Makefile`              # Build automation (see ADR-004)
- `docker-compose.dev.yml` # Local development environment
- `go.mod`, `go.sum`      # Go module dependencies
- `package.json`          # Frontend dependencies
- `CLAUDE.md`             # AI assistant guidance
```

## Purpose by Directory

### backend/
Go API server implementing compose operations, RBAC, and storage abstraction.
- Uses Gin for HTTP, GORM for PostgreSQL
- Storage interface switches between S3 (prod) and Local FS (dev)
- JWT middleware enforces Developer/Engineer roles

### frontend/
React TypeScript UI for compose management and offline packaging.
- Tailwind CSS for styling
- Role-based UI (Developer can modify, Engineer read-only)

### ops/
Operational configuration and documentation.
- ADRs define architectural constraints (no `build:`, prebuilt images only)
- Compose templates define module structure for merging

### tools/
CLI implementations of core functionality.
- Stateless tools for CI/CD integration
- Can be used standalone or via API

### test/
Testing infrastructure prioritizing compose merge/lint/packaging.
- Integration tests validate multi-module scenarios
- Fixtures provide reproducible test cases

### dist/
Build outputs (not committed).
- Docker images saved as .tar for offline use
- Complete installer packages with scripts

## References
- ADR-001: Compose merge with namespacing
- ADR-002: Lint policy enforcement
- ADR-003: Offline packaging format
- ADR-004: Repository workflow