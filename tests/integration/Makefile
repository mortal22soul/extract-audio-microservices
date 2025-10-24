# Integration Testing Makefile

.PHONY: help setup test test-grpc test-workflows test-realtime test-chaos cleanup coverage

# Default target
help:
	@echo "Available targets:"
	@echo "  setup       - Set up test environment"
	@echo "  test        - Run all integration tests"
	@echo "  test-grpc   - Run gRPC service tests"
	@echo "  test-workflows - Run workflow tests"
	@echo "  test-realtime - Run real-time tests"
	@echo "  test-chaos  - Run chaos engineering tests"
	@echo "  coverage    - Run tests with coverage"
	@echo "  cleanup     - Clean up test environment"

# Set up test environment
setup:
	@echo "Setting up integration test environment..."
	docker-compose -f docker-compose.test.yml up -d
	@echo "Waiting for services to be ready..."
	sleep 30
	@echo "Test environment ready!"

# Run all integration tests
test: setup
	@echo "Running all integration tests..."
	go test ./... -v -timeout=10m
	$(MAKE) cleanup

# Run gRPC service tests
test-grpc: setup
	@echo "Running gRPC service tests..."
	go test ./grpc -v -timeout=5m
	$(MAKE) cleanup

# Run workflow tests
test-workflows: setup
	@echo "Running workflow tests..."
	go test ./workflows -v -timeout=10m
	$(MAKE) cleanup

# Run real-time tests
test-realtime: setup
	@echo "Running real-time tests..."
	go test ./realtime -v -timeout=5m
	$(MAKE) cleanup

# Run chaos engineering tests
test-chaos: setup
	@echo "Running chaos engineering tests..."
	go test ./chaos -v -timeout=15m
	$(MAKE) cleanup

# Run tests with coverage
coverage: setup
	@echo "Running integration tests with coverage..."
	go test ./... -v -timeout=10m -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"
	$(MAKE) cleanup

# Clean up test environment
cleanup:
	@echo "Cleaning up test environment..."
	docker-compose -f docker-compose.test.yml down -v
	docker system prune -f
	@echo "Test environment cleaned up!"

# Quick test (without full setup/cleanup)
test-quick:
	@echo "Running quick integration tests (assumes environment is ready)..."
	go test ./... -v -timeout=5m

# Test specific package
test-package:
	@if [ -z "$(PKG)" ]; then \
		echo "Usage: make test-package PKG=<package_name>"; \
		echo "Example: make test-package PKG=grpc"; \
		exit 1; \
	fi
	@echo "Running tests for package: $(PKG)"
	go test ./$(PKG) -v -timeout=5m

# Lint integration tests
lint:
	@echo "Linting integration tests..."
	golangci-lint run ./...

# Format integration tests
fmt:
	@echo "Formatting integration tests..."
	go fmt ./...

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy

# Build test binaries
build:
	@echo "Building test binaries..."
	go test -c ./grpc -o bin/grpc.test
	go test -c ./workflows -o bin/workflows.test
	go test -c ./realtime -o bin/realtime.test
	go test -c ./chaos -o bin/chaos.test

# Run tests in CI environment
test-ci:
	@echo "Running integration tests in CI environment..."
	# Set CI-specific environment variables
	export TEST_POSTGRES_URL="postgres://test_user:test_pass@postgres-test:5432/video_converter_test?sslmode=disable"
	export TEST_MONGO_URL="mongodb://test_user:test_pass@mongodb-test:27017/video_converter_test"
	export TEST_REDIS_URL="redis://:test_pass@redis-test:6379"
	export TEST_RABBITMQ_URL="amqp://test_user:test_pass@rabbitmq-test:5672/"
	export TEST_AUTH_SERVICE_URL="auth-service-test:50051"
	export TEST_ANALYTICS_URL="analytics-service-test:50052"
	export TEST_GATEWAY_URL="http://gateway-service-test:8080"
	export TEST_REALTIME_URL="http://realtime-service-test:3001"
	go test ./... -v -timeout=15m -coverprofile=coverage.out

# Generate test report
report:
	@echo "Generating test report..."
	go test ./... -v -json > test-results.json
	@echo "Test report generated: test-results.json"