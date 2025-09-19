# Burndler Makefile
# Enforces: NO build: directives, prebuilt images only, image@sha256 preferred

# Version information
VERSION ?= $(shell cat VERSION 2>/dev/null || echo "dev")
BUILD_TIME ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
GIT_COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS = -X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT)

# Initialization marker
INIT_MARKER := .initialized

.PHONY: help init check-init install-golangci-lint dev dev-backend dev-frontend build build-docker test lint clean version release

help: ## Show this help message
	@echo "Burndler Development Commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# ===== Initialization =====

init: ## Initialize development environment (install all required tools)
	@echo "🔧 Initializing Burndler development environment..."
	@make deps-backend
	@make deps-frontend
	@make install-golangci-lint
	@touch $(INIT_MARKER)
	@echo "✅ Development environment initialized successfully!"
	@echo "You can now run 'make dev', 'make test', or 'make build'"

check-init: ## Check if development environment is initialized
	@if [ ! -f $(INIT_MARKER) ]; then \
		echo "⚠️  Warning: Development environment not initialized!"; \
		echo "Please run 'make init' first to install required tools."; \
		echo ""; \
		exit 1; \
	fi

install-golangci-lint: ## Install golangci-lint tool
	@echo "📦 Installing golangci-lint..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		echo "✓ golangci-lint already installed"; \
	else \
		echo "Installing golangci-lint..."; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin latest; \
		echo "✓ golangci-lint installed"; \
	fi

# ===== Development =====

dev: dev-db ## Start full development environment (backend + frontend in parallel)
	@echo "🚀 Starting Burndler Development Environment..."
	@echo "📦 Database is ready!"
	@echo "🔄 Starting backend and frontend in parallel..."
	@echo ""
	@echo "  🌐 Backend API:  http://localhost:8080"
	@echo "  🌐 Frontend:     http://localhost:3000"
	@echo "  🗄️  PostgreSQL:   localhost:5432"
	@echo ""
	@echo "Press Ctrl+C to stop all services"
	@echo ""
	@make -j 2 dev-backend dev-frontend

dev-backend: ## Start backend with Air hot reload (requires PostgreSQL)
	@echo "🔧 Starting backend development with Air hot reload..."
	@cp -n .env.example .env.development 2>/dev/null || true
	@make ensure-dev-db
	@echo "✅ Database confirmed running"
	@mkdir -p tmp
	@echo "🔥 Starting Air hot reload..."
	@if command -v air >/dev/null 2>&1; then \
		cd backend && air -c ../.air.toml; \
	else \
		echo "❌ Air not installed. Installing..."; \
		go install github.com/cosmtrek/air@latest; \
		cd backend && air -c ../.air.toml; \
	fi

backend: dev-backend ## Alias for dev-backend (start backend with Air hot reload)

dev-frontend: ## Start frontend with Vite dev server
	@echo "⚡ Starting frontend development with Vite..."
	@echo "🌐 Frontend will be available at: http://localhost:3000"
	@echo "🔄 Hot Module Replacement enabled"
	cd frontend && npm run dev

dev-db: ## Start PostgreSQL database only
	@echo "🗄️  Starting PostgreSQL database for development..."
	@cp -n .env.example .env.development 2>/dev/null || true
	docker-compose -f compose/postgres.compose.yaml --env-file .env.development up -d
	@echo "✅ PostgreSQL started on localhost:5432"
	@echo "   📋 Database: burndler_dev"
	@echo "   📋 Test DB:  burndler_test"
	@echo "   👤 User:     burndler"

ensure-dev-db: ## Ensure development database is running
	@if ! docker ps | grep burndler_postgres_dev > /dev/null; then \
		echo "📦 PostgreSQL not running, starting..."; \
		make dev-db; \
		echo "⏳ Waiting for database to be ready..."; \
		sleep 5; \
	fi
	@docker exec burndler_postgres_dev pg_isready -U burndler -d burndler_dev > /dev/null 2>&1 || (echo "⏳ Waiting for PostgreSQL..." && sleep 3)

dev-reset: ## Reset entire development environment
	@echo "🔄 Resetting development environment..."
	@make dev-down
	@echo "🧹 Cleaning up development data..."
	docker-compose -f compose/postgres.compose.yaml --env-file .env.development down -v
	@echo "🚀 Restarting fresh environment..."
	@make dev-db
	@echo "✅ Environment reset complete!"

dev-down: ## Stop all development services
	@echo "🛑 Stopping development services..."
	docker-compose -f compose/postgres.compose.yaml --env-file .env.development down
	@echo "✅ Development services stopped"


dev-logs: ## Show development database logs
	@echo "📋 PostgreSQL logs:"
	docker-compose -f compose/postgres.compose.yaml --env-file .env.development logs -f postgres



dev-status: ## Show development services status
	@echo "📊 Development Services Status:"
	@echo ""
	@if docker ps | grep burndler_postgres_dev > /dev/null; then \
		echo "✅ PostgreSQL: Running on localhost:5432"; \
	else \
		echo "❌ PostgreSQL: Not running"; \
	fi
	@echo ""
	@if curl -s http://localhost:8080/health > /dev/null 2>&1; then \
		echo "✅ Backend API: Running on localhost:8080"; \
	else \
		echo "❌ Backend API: Not running"; \
	fi
	@echo ""
	@if curl -s http://localhost:3000 > /dev/null 2>&1; then \
		echo "✅ Frontend: Running on localhost:3000"; \
	else \
		echo "❌ Frontend: Not running"; \
	fi

# ===== Build =====

build: check-init build-backend-with-static ## Build all components

build-backend: ## Build Go binary
	@echo "Building backend binary with version $(VERSION)..."
	cd backend && go build -ldflags="$(LDFLAGS)" -o ../dist/burndler cmd/api/main.go

build-frontend: ## Build React production bundle
	@echo "Building frontend bundle..."
	cd frontend && npm run build

prepare-static: build-frontend ## Copy frontend build to backend for embedding
	@echo "Preparing static files for embedding..."
	@rm -rf backend/internal/static/dist
	@cp -r frontend/dist backend/internal/static/dist

build-backend-with-static: prepare-static ## Build Go binary with embedded frontend
	@echo "Building backend binary with embedded frontend (v$(VERSION))..."
	cd backend && go build -ldflags="$(LDFLAGS)" -o ../dist/burndler cmd/api/main.go

build-docker: ## Build Docker image with embedded frontend
	@echo "Building Docker image v$(VERSION)..."
	docker build \
		--build-arg VERSION=$(VERSION) \
		--build-arg BUILD_TIME=$(BUILD_TIME) \
		--build-arg GIT_COMMIT=$(GIT_COMMIT) \
		-t burndler:latest \
		-t burndler:$(VERSION) \
		.
	@echo "Docker images built: burndler:latest and burndler:$(VERSION)"


# ===== Testing =====

test: check-init test-unit test-integration ## Run all tests

test-unit: ## Run unit tests
	@echo "Running unit tests..."
	cd backend && go test -v -short ./...
	cd frontend && npm test

test-integration: ## Run integration tests (requires database)
	@echo "Running integration tests..."
	cd backend && go test -v ./...
	cd test/integration && go test ./...

test-e2e: ## Run end-to-end tests with Playwright
	@echo "Running E2E tests..."
	cd frontend && npm run test:e2e

test-coverage: ## Generate test coverage report
	@echo "Generating coverage report..."
	cd backend && go test -coverprofile=coverage.out ./...
	cd backend && go tool cover -html=coverage.out -o coverage.html

# ===== Quality =====

format: format-backend format-frontend ## Format all code to match CI requirements

format-backend: ## Fix Go code issues with golangci-lint
	@echo "🔧 Fixing Go code issues..."
	@make install-golangci-lint
	cd backend && golangci-lint run --fix
	@echo "✅ Go code checked and fixed"

format-frontend: ## Format frontend code with Prettier
	@echo "🎨 Formatting frontend code..."
	cd frontend && npm run format
	@echo "✅ Frontend code formatted"

pre-commit: format-backend format-frontend ## Run all pre-commit formatting
	@echo "🔍 Verifying CI compliance..."
	@make install-golangci-lint
	@make lint-backend
	@make lint-frontend
	@echo "✅ Ready to commit! All CI checks will pass"

lint: lint-backend lint-frontend lint-compose ## Run all linters

lint-backend: ## Run golangci-lint
	@echo "Linting Go code..."
	@make install-golangci-lint
	cd backend && golangci-lint run

lint-frontend: ## Run ESLint and Prettier
	@echo "Linting JavaScript/TypeScript..."
	cd frontend && npm run lint
	cd frontend && npm run format:check

lint-compose: ## Validate compose files (no build:, etc.)
	@echo "Validating compose files..."
	@echo "Checking for forbidden 'build:' directives..."
	@! grep -r "^[[:space:]]*build:" compose/*.yaml 2>/dev/null || (echo "ERROR: 'build:' directive found in compose files!" && exit 1)
	@echo "Checking for image SHA256 usage..."
	@grep -E "image:.*@sha256:" compose/dev.compose.yaml > /dev/null || echo "WARNING: Not all images use SHA256 digests"
	@echo "Compose files validated ✓"

# ===== Operations =====


init-backend: ## Initialize Go module (run once)
	cd backend && go mod init github.com/burndler/burndler

init-frontend: ## Initialize frontend with npm (run once)
	cd frontend && npm init -y && npm install

deps-backend: ## Download Go dependencies
	cd backend && go mod download

deps-frontend: ## Install frontend dependencies
	cd frontend && npm install

# ===== Docker Operations =====

docker-lint: ## Lint Dockerfiles (if any)
	@echo "No Dockerfiles should exist (prebuilt images only)"
	@! find . -name "Dockerfile*" -not -path "./.git/*" | grep . || (echo "ERROR: Dockerfiles found! Use prebuilt images only" && exit 1)

image-list: ## List all images from compose files
	@echo "Images used in compose files:"
	@grep -h "image:" compose/*.yaml | sed 's/.*image: *//' | sort -u

# ===== Utility =====

clean: ## Remove build artifacts
	@echo "Cleaning build artifacts..."
	rm -rf dist/*
	rm -rf backend/coverage.*
	rm -rf backend/internal/static/dist
	rm -rf frontend/dist
	rm -rf frontend/coverage
	find . -name "*.log" -delete

setup: ## Initial project setup
	@echo "Setting up project..."
	@cp -n .env.example .env || true
	@make init-backend
	@make init-frontend
	@make deps-backend
	@make deps-frontend
	@echo "Setup complete! Run 'make dev' to start development"

verify: ## Verify project constraints
	@echo "Verifying project constraints..."
	@make lint-compose
	@make docker-lint
	@echo "✓ No build: directives found"
	@echo "✓ No Dockerfiles found"
	@echo "✓ Using prebuilt images only"
	@echo "Project constraints verified!"

# ===== CI/CD Helpers =====

ci-validate: lint-compose docker-lint ## CI validation step
	@echo "CI validation passed"

ci-test: test-unit test-integration ## CI test step
	@echo "CI tests passed"

ci-build: build ## CI build step
	@echo "CI build completed"

# ===== Release Management =====

version: ## Show current version
	@echo "Burndler v$(VERSION)"
	@echo "Build Time: $(BUILD_TIME)"
	@echo "Git Commit: $(GIT_COMMIT)"

release-local: ## Prepare local release build
	@echo "Preparing local release v$(VERSION)..."
	@make clean
	@make build
	@make build-docker
	@echo "✅ Local release v$(VERSION) ready"

release-check: ## Validate release readiness
	@echo "🔍 Checking release readiness..."
	@if [ "$(VERSION)" = "dev" ]; then \
		echo "❌ VERSION file must contain a proper version (not 'dev')"; \
		exit 1; \
	fi
	@if ! git diff-index --quiet HEAD --; then \
		echo "❌ Working directory not clean. Commit or stash changes."; \
		exit 1; \
	fi
	@if [ -z "$(shell git tag -l "v$(VERSION)")" ]; then \
		echo "✅ Version v$(VERSION) is not yet tagged"; \
	else \
		echo "❌ Version v$(VERSION) already tagged"; \
		exit 1; \
	fi
	@echo "✅ Release readiness check passed"

pre-commit-install: ## Install pre-commit hooks
	@echo "Installing pre-commit hooks..."
	@which pre-commit > /dev/null || pip install pre-commit
	@pre-commit install
	@echo "✅ Pre-commit hooks installed"

pre-commit-run: ## Run pre-commit hooks on all files
	@echo "Running pre-commit hooks..."
	@pre-commit run --all-files

sync-version: ## Sync version across package.json and VERSION file
	@echo "Syncing version $(VERSION) across files..."
	@sed -i 's/"version": "[^"]*"/"version": "$(VERSION)"/' frontend/package.json
	@echo "✅ Version synced to $(VERSION)"