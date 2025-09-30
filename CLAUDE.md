# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Architecture

Burndler is a module-based Docker Compose orchestration platform that enables users to create, version, and combine reusable deployment modules into projects for offline deployment. It provides a complete module registry and project management system for containerized applications.

## Core Concepts

### Module
A **module** is a reusable deployment unit consisting of:
- A `docker-compose.yaml` file defining services
- Associated deployment resources (configs, scripts, templates)
- Version management capabilities
- Published to a centralized module registry

### Module Version
Each module can have multiple **versions** (e.g., v0.1.0, v0.1.1):
- Semantic versioning support
- Immutable once published
- Contains versioned compose content and resources
- Enables dependency management and compatibility

### Project
A **project** is a combination of multiple modules:
- Selects specific module versions to include
- Provides project-level variable overrides
- Defines module composition and ordering
- Builds into complete deployment packages

### Installer
The final **installer** package enables complete offline deployment:
- Merged docker-compose.yaml from all project modules
- All required Docker images as .tar files
- Module resources and configuration files
- Installation and verification scripts
- Complete offline deployment capability

## Module and Project Workflow

1. **Module Creation**: Users define new modules or import existing ones
2. **Module Versioning**: Modules are versioned and published to registry
3. **Project Composition**: Users select and combine module versions into projects
4. **Project Building**: System merges modules, resolves images, creates installer
5. **Offline Deployment**: Installer contains everything needed for air-gapped deployment

### Core Services
- **module-service**: Manages module registry, versioning, and CRUD operations
- **project-service**: Handles project creation, module composition, and configuration
- **image-service**: Resolves Docker images, pulls from registries, packages as .tar files
- **compose-merger**: Merges module docker-compose.yml files with namespace prefixing (`{namespace}__{name}`)
- **compose-linter**: Validates merged compose files against security and configuration policies
- **build-service**: Orchestrates project builds from modules to complete installers
- **installer-packager**: Creates offline installer packages with `install.sh`/`verify.sh` scripts
- **rbac-security**: Implements JWT middleware with Developer(RW)/Engineer(R) role enforcement

### Technology Stack
- **Backend**: Go 1.24 + Gin framework, GORM ORM, PostgreSQL database
- **Frontend**: Node.js 20 + React + TypeScript + Tailwind CSS
- **Storage**: S3-compatible (production) or Local FS (dev/offline), interface-switchable via environment
- **Deployment**: Docker Compose with prebuilt images only

## Development Commands

Standard development workflow:
- `make dev` - Start full development environment (backend + frontend + postgres)
- `make dev-backend` - Start backend only (Go + PostgreSQL)
- `make dev-frontend` - Start frontend only (React dev server)
- `make test-unit` - Run unit tests (no database required)
- `make test-integration` - Run integration tests (requires database)
- `make build` - Build all components for production

### ⚠️ Make Command Execution Rule

**ALL `make` commands MUST be executed from project root**.

The Makefile exists ONLY at project root. Always run:
```bash
cd <Project Root Directory> && make <command>
```

## Architectural Rules & Constraints

### Compose/Packaging (STRICT)
- **Prebuilt images only**: `build:` directive is forbidden and will cause lint failures
- **Image references**: Prefer `image@sha256:...` format for reproducibility
- **Namespacing**: All services/networks/volumes prefixed as `{namespace}__{name}`
- **Variable substitution**: Project-level variables override module defaults

## Database Entities

### Core Models
- **Module**: Registry of reusable compose modules with metadata
- **ModuleVersion**: Versioned releases of modules (immutable once published)
- **Project**: Collection of modules for deployment with configuration
- **ProjectModule**: Many-to-many relationship with version pinning and overrides
- **Build**: Project build instances creating installer packages
- **User**: Authentication and authorization for module/project access
- **Setup**: System initialization and configuration state

### Entity Relationships
```
User 1:N Project 1:N ProjectModule N:1 ModuleVersion N:1 Module
User 1:N Build N:1 Project
```

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
1. **Module management** - CRUD operations, versioning, and publication
2. **Project composition** - module selection, variable overrides, dependency resolution
3. **Build pipeline** - project-to-installer transformation with image packaging
4. **Compose merge functionality** - namespace handling, variable substitution
5. **Lint policy enforcement** - security rules, reference validation
6. **Docker image packaging** - registry resolution, image pulling, tar creation
7. **RBAC implementation** - JWT middleware and role enforcement

## ADR References
- **ADR-001**: Module registry and versioning strategy
- **ADR-002**: Project composition and dependency management
- **ADR-003**: Compose merge strategy with namespacing and conflict resolution
- **ADR-004**: Lint policy covering security, variables, and references
- **ADR-005**: Docker image packaging and offline installer format
- **ADR-006**: Storage architecture for modules, projects, and builds

## Conflict Resolution
If a request conflicts with these architectural rules:
1. **Warn** about the conflict with specific rule reference
2. **Propose** compliant alternative approach
3. **Abort** if user insists on non-compliant change

## Pre-commit Formatting Requirements (MANDATORY)

Claude MUST run these exact commands before ANY commit to match CI requirements:

### Backend (Go)
```bash
make format-backend
make lint-backend
```

### Frontend (React/TypeScript)
```bash
make format-frontend
make lint-frontend
```

### Commit Checklist
1. ✅ Backend: `make format-backend && make lint-backend` passes without errors
2. ✅ Frontend: `make format-frontend && make lint-frontend` passes (no formatting issues)
3. ✅ Include any auto-fixed changes in the commit
4. ✅ If formatting modified files, mention in a commit message

## API Structure

### Module Management
```
GET    /api/v1/modules                    # List modules with pagination
POST   /api/v1/modules                    # Create module (Developer only)
GET    /api/v1/modules/{id}               # Get module details
PUT    /api/v1/modules/{id}               # Update module (Developer only)
DELETE /api/v1/modules/{id}               # Delete module (Developer only)

GET    /api/v1/modules/{id}/versions      # List module versions
POST   /api/v1/modules/{id}/versions      # Create version (Developer only)
GET    /api/v1/modules/{id}/versions/{version} # Get specific version
PUT    /api/v1/modules/{id}/versions/{version} # Update version (Developer only)
POST   /api/v1/modules/{id}/versions/{version}/publish # Publish version
```

### Project Management
```
GET    /api/v1/projects                   # List user projects
POST   /api/v1/projects                   # Create project
GET    /api/v1/projects/{id}              # Get project details
PUT    /api/v1/projects/{id}              # Update project
DELETE /api/v1/projects/{id}              # Delete project

GET    /api/v1/projects/{id}/modules      # List project modules
POST   /api/v1/projects/{id}/modules      # Add module to project
PUT    /api/v1/projects/{id}/modules/{module_id} # Update module config
DELETE /api/v1/projects/{id}/modules/{module_id} # Remove module

POST   /api/v1/projects/{id}/validate     # Validate project composition
POST   /api/v1/projects/{id}/build        # Build installer package
```

### Enhanced Build Management
```
GET    /api/v1/builds                     # List builds with filters
POST   /api/v1/builds                     # Create build (direct or project-based)
GET    /api/v1/builds/{id}                # Get build status
DELETE /api/v1/builds/{id}                # Cancel/delete build
GET    /api/v1/builds/{id}/download       # Download installer package
GET    /api/v1/builds/{id}/logs           # Get build logs
```

## Programming Instructions
Always adhere to @.claude/docs/TDD.md when writing or modifying source code.
