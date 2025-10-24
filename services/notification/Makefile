# Notification Service Makefile

.PHONY: build run test clean deps lint fmt vet docker-build docker-run help

# Variables
BINARY_NAME=notification
DOCKER_IMAGE=notification-service
GO_FILES=$(shell find . -name "*.go" -type f)

# Default target
help: ## Show this help message
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-15s %s\n", $$1, $$2}'

# Development
deps: ## Install dependencies
	go mod download
	go mod tidy

build: ## Build the binary
	go build -o $(BINARY_NAME) cmd/main.go

run: ## Run the service
	go run cmd/main.go

dev: ## Run with auto-reload (requires air)
	air

# Testing
test: ## Run tests
	go test -v ./...

test-coverage: ## Run tests with coverage
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Code quality
lint: ## Run linter
	golangci-lint run

fmt: ## Format code
	go fmt ./...

vet: ## Run go vet
	go vet ./...

# Docker
docker-build: ## Build Docker image
	docker build -f Dockerfile.dev -t $(DOCKER_IMAGE) .

docker-run: ## Run Docker container
	docker run -p 8080:8080 --env-file .env $(DOCKER_IMAGE)

# Utilities
clean: ## Clean build artifacts
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html

# Test endpoints (requires service to be running)
test-welcome: ## Send test welcome email
	curl -X POST http://localhost:8080/test/welcome \
		-H "Content-Type: application/json" \
		-d '{"user_id": "test123", "email": "test@example.com", "first_name": "Test User"}'

test-conversion-complete: ## Send test conversion complete email
	curl -X POST http://localhost:8080/test/conversion-complete \
		-H "Content-Type: application/json" \
		-d '{"user_id": "test123", "email": "test@example.com", "first_name": "Test User", "video_name": "test-video.mp4", "job_id": "job123", "duration": "3:45", "file_size": "15.2 MB", "conversion_time": "1m 30s"}'

test-conversion-error: ## Send test conversion error email
	curl -X POST http://localhost:8080/test/conversion-error \
		-H "Content-Type: application/json" \
		-d '{"user_id": "test123", "email": "test@example.com", "first_name": "Test User", "video_name": "test-video.mp4", "job_id": "job123", "error_message": "Unsupported video format"}'

health: ## Check service health
	curl http://localhost:8080/health

# Development setup
setup: deps ## Setup development environment
	@echo "Setting up development environment..."
	@if [ ! -f .env ]; then cp .env.example .env; echo "Created .env file - please configure your settings"; fi
	@echo "Development environment ready!"
	@echo "1. Configure your .env file"
	@echo "2. Run 'make run' to start the service"
	@echo "3. Run 'make test-welcome' to send a test email"