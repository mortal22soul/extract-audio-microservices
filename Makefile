.PHONY: help dev build test clean proto tilt-up tilt-down docker-up docker-down

# Default target
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Development
dev: ## Start all services in development mode with Tilt
	tilt up

build: ## Build all services
	@echo "Building Go services..."
	@cd services/gateway && go build -o bin/gateway ./cmd/main.go
	@cd services/auth && go build -o bin/auth ./cmd/main.go
	@cd services/converter && go build -o bin/converter ./cmd/main.go
	@cd services/notification && go build -o bin/notification ./cmd/main.go
	@echo "Building TypeScript services..."
	@pnpm build
	@echo "Building Python service..."
	@cd services/analytics && uv sync

test: ## Run tests for all services
	@echo "Running Go tests..."
	@cd services/gateway && go test ./...
	@cd services/auth && go test ./...
	@cd services/converter && go test ./...
	@cd services/notification && go test ./...
	@echo "Running TypeScript tests..."
	@pnpm test
	@echo "Running Python tests..."
	@cd services/analytics && uv run pytest

clean: ## Clean build artifacts
	@echo "Cleaning Go binaries..."
	@find services -name "bin" -type d -exec rm -rf {} +
	@echo "Cleaning TypeScript builds..."
	@pnpm clean
	@echo "Cleaning Python cache..."
	@find services/analytics -name "__pycache__" -type d -exec rm -rf {} +
	@find services/analytics -name "*.pyc" -delete

# Protocol Buffers
proto: ## Generate protobuf code for all languages
	buf generate

# Tilt.dev
tilt-up: ## Start Tilt development environment
	tilt up

tilt-down: ## Stop Tilt development environment
	tilt down

# Docker Compose (fallback)
docker-up: ## Start services with Docker Compose
	docker-compose up -d

docker-down: ## Stop Docker Compose services
	docker-compose down

# Dependencies
deps-go: ## Download Go dependencies
	@echo "Installing Go dependencies..."
	@cd services/gateway && go mod download && go mod tidy
	@cd services/auth && go mod download && go mod tidy
	@cd services/converter && go mod download && go mod tidy
	@cd services/notification && go mod download && go mod tidy

deps-node: ## Install Node.js dependencies
	@echo "Installing Node.js dependencies..."
	@pnpm install

deps-python: ## Install Python dependencies
	@echo "Installing Python dependencies..."
	@cd services/analytics && uv sync --dev

deps: deps-go deps-node deps-python ## Install all dependencies

# Development tools setup
setup-dev: deps pre-commit-install ## Set up development environment
	@echo "Development environment setup complete!"
	@echo "Run 'make dev' to start all services with Tilt"
	@echo "Run 'make docker-up' to start with Docker Compose"

# Code quality checks
check: lint test ## Run all code quality checks

check-go: ## Run Go-specific checks
	@echo "Running Go checks..."
	@golangci-lint run ./services/gateway/...
	@golangci-lint run ./services/auth/...
	@golangci-lint run ./services/converter/...
	@golangci-lint run ./services/notification/...
	@cd services/gateway && go test ./...
	@cd services/auth && go test ./...
	@cd services/converter && go test ./...
	@cd services/notification && go test ./...

check-ts: ## Run TypeScript-specific checks
	@echo "Running TypeScript checks..."
	@pnpm lint
	@pnpm type-check
	@pnpm test

check-python: ## Run Python-specific checks
	@echo "Running Python checks..."
	@cd services/analytics && uv run ruff check .
	@cd services/analytics && uv run mypy src/
	@cd services/analytics && uv run pytest

# Linting and Formatting
lint: lint-go lint-ts lint-python ## Run linters for all services

lint-go: ## Lint Go code
	@echo "Linting Go code..."
	@golangci-lint run ./services/gateway/...
	@golangci-lint run ./services/auth/...
	@golangci-lint run ./services/converter/...
	@golangci-lint run ./services/notification/...

lint-ts: ## Lint TypeScript code
	@echo "Linting TypeScript code..."
	@pnpm lint

lint-python: ## Lint Python code
	@echo "Linting Python code..."
	@cd services/analytics && uv run ruff check .
	@cd services/analytics && uv run mypy src/

format: format-go format-ts format-python ## Format code for all services

format-go: ## Format Go code
	@echo "Formatting Go code..."
	@gofmt -w services/gateway/ services/auth/ services/converter/ services/notification/
	@goimports -w services/gateway/ services/auth/ services/converter/ services/notification/

format-ts: ## Format TypeScript code
	@echo "Formatting TypeScript code..."
	@pnpm prettier --write "services/**/*.{ts,tsx,js,jsx,json,md}"

format-python: ## Format Python code
	@echo "Formatting Python code..."
	@cd services/analytics && uv run black .
	@cd services/analytics && uv run ruff --fix .

lint-fix: ## Auto-fix linting issues where possible
	@echo "Auto-fixing Go imports..."
	@goimports -w services/gateway/ services/auth/ services/converter/ services/notification/
	@echo "Auto-fixing TypeScript issues..."
	@pnpm eslint --fix "services/**/*.{ts,tsx}"
	@echo "Auto-fixing Python issues..."
	@cd services/analytics && uv run ruff --fix .

# Pre-commit hooks
pre-commit-install: ## Install pre-commit hooks
	@echo "Installing pre-commit hooks..."
	@pip install pre-commit
	@pre-commit install
	@pre-commit install --hook-type commit-msg

pre-commit-run: ## Run pre-commit hooks on all files
	@echo "Running pre-commit hooks..."
	@pre-commit run --all-files

pre-commit-update: ## Update pre-commit hooks
	@echo "Updating pre-commit hooks..."
	@pre-commit autoupdate