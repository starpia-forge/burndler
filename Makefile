# Burndler Makefile
# Enforces: NO build: directives, prebuilt images only, image@sha256 preferred

# Version information
VERSION ?= $(shell cat VERSION 2>/dev/null || echo "dev")
BUILD_TIME ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
GIT_COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS = -X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -X main.GitCommit=$(GIT_COMMIT)

.PHONY: help dev dev-backend dev-frontend build build-docker test lint clean version release

help: ## Show this help message
	@echo "Burndler Development Commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# ===== Development =====

dev: ## Start full development environment (backend + frontend + postgres)
	@echo "Starting development environment..."
	@cp -n .env.example .env 2>/dev/null || true
	docker-compose -f compose/dev.compose.yaml up

dev-backend: ## Start backend only (Go + PostgreSQL)
	@echo "Starting backend development..."
	@cp -n .env.example .env 2>/dev/null || true
	docker-compose -f compose/dev.compose.yaml up postgres backend

dev-frontend: ## Start frontend only (React dev server)
	@echo "Starting frontend development..."
	docker-compose -f compose/dev.compose.yaml up frontend

dev-down: ## Stop all development containers
	docker-compose -f compose/dev.compose.yaml down

dev-clean: ## Stop and remove all dev containers and volumes
	docker-compose -f compose/dev.compose.yaml down -v

# ===== Build =====

build: build-backend-with-static build-tools ## Build all components

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

build-tools: ## Build CLI tools
	@echo "Building CLI tools v$(VERSION)..."
	cd tools/merge && go build -ldflags="-X main.Version=$(VERSION)" -o ../../dist/burndler-merge .
	cd tools/lint && go build -ldflags="-X main.Version=$(VERSION)" -o ../../dist/burndler-lint .
	cd tools/package && go build -ldflags="-X main.Version=$(VERSION)" -o ../../dist/burndler-package .

# ===== Testing =====

test: test-unit test-integration ## Run all tests

test-unit: ## Run unit tests
	@echo "Running unit tests..."
	cd backend && go test ./...
	cd frontend && npm test

test-integration: ## Run integration tests
	@echo "Running integration tests..."
	cd test/integration && go test ./...

test-e2e: ## Run end-to-end tests with Playwright
	@echo "Running E2E tests..."
	cd frontend && npm run test:e2e

test-coverage: ## Generate test coverage report
	@echo "Generating coverage report..."
	cd backend && go test -coverprofile=coverage.out ./...
	cd backend && go tool cover -html=coverage.out -o coverage.html

# ===== Quality =====

lint: lint-go lint-js lint-compose ## Run all linters

lint-go: ## Run golangci-lint
	@echo "Linting Go code..."
	@which golangci-lint > /dev/null || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	cd backend && golangci-lint run
	cd tools && golangci-lint run

lint-js: ## Run ESLint and Prettier
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

merge: ## Test compose merge functionality
	@echo "Testing compose merge..."
	go run tools/merge/main.go --namespace test --input test/fixtures/compose/module1.yaml --input test/fixtures/compose/module2.yaml

package: ## Create offline installer package
	@echo "Creating offline installer..."
	go run tools/package/main.go --compose compose/dev.compose.yaml --output dist/installers/

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