# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Architecture

Burndler is a Docker Compose orchestration tool that merges, lints, and packages multi-module Docker applications for offline deployment. It consists of specialized agents that handle different aspects of the deployment pipeline.

### Core Agents
- **compose-merger**: Merges multiple docker-compose.yml files with namespace prefixing (`{namespace}__{name}`) and variable substitution
- **compose-linter**: Validates merged compose files against security and configuration policies
- **image-packager**: Resolves image tags to digests, pulls and packages into `.tar` files for offline use
- **installer-packager**: Creates offline installer packages with `install.sh`/`verify.sh` scripts
- **rbac-security**: Implements JWT middleware with Developer(RW)/Engineer(R) role enforcement
- **architect**: Manages ADRs and repository conventions

### Technology Stack
- **Backend**: Go 1.22 + Gin framework, GORM ORM, PostgreSQL database
- **Frontend**: React + TypeScript + Tailwind CSS
- **Storage**: S3-compatible (production) or Local FS (dev/offline), interface-switchable via environment
- **Deployment**: Docker Compose with prebuilt images only

## Development Commands

Since this appears to be early-stage project, check for:
- `go run main.go` or equivalent for backend development
- `npm run dev` or `yarn dev` for frontend development
- `docker-compose up -d` for local environment

## Architectural Rules & Constraints

### Compose/Packaging (STRICT)
- **Prebuilt images only**: `build:` directive is forbidden and will cause lint failures
- **Image references**: Prefer `image@sha256:...` format for reproducibility
- **Namespacing**: All services/networks/volumes prefixed as `{namespace}__{name}`
- **Variable substitution**: Project-level variables override module defaults

### Security & RBAC
- **Roles**: Developer (full access), Engineer (read-only)
- **Enforcement**: Route-level RBAC checks with JWT middleware
- **Secrets**: Never commit secrets; use environment files or Docker secrets

### Linting Rules (MUST PASS)
- Unresolved environment variables → ERROR
- Host port collisions → ERROR
- Any `build:` directive → ERROR
- Invalid service/volume/network references → ERROR
- Privileged containers or cap_add → WARNING

### Offline Installer Structure
```
installer.tar.gz/
├── compose/docker-compose.yaml     # Merged compose file
├── images/*.tar                    # Docker images (deduplicated by digest)
├── resources/<module>/<version>/   # Static resources
├── env/.env.example               # Environment template
├── bin/install.sh                 # Installation script
├── bin/verify.sh                  # Verification script
└── manifest.json                  # Package metadata
```

## Testing Priorities
1. **Compose merge functionality** - namespace handling, variable substitution
2. **Lint policy enforcement** - security rules, reference validation
3. **Packaging pipeline** - offline installer generation and verification
4. **RBAC implementation** - JWT middleware and role enforcement

## ADR References
- **ADR-001**: Compose merge strategy with namespacing and conflict resolution
- **ADR-002**: Lint policy covering security, variables, and references
- **ADR-003**: Packaging format for reproducible offline installers

## Conflict Resolution
If a request conflicts with these architectural rules:
1. **Warn** about the conflict with specific rule reference
2. **Propose** compliant alternative approach
3. **Abort** if user insists on non-compliant change