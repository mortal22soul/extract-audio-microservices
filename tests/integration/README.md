# Integration Testing Framework

This directory contains comprehensive integration tests for the microservices architecture.

## Test Structure

- `grpc/` - gRPC service communication tests
- `workflows/` - End-to-end workflow tests
- `realtime/` - Real-time notification tests
- `chaos/` - Chaos engineering tests
- `utils/` - Test utilities and helpers
- `docker-compose.test.yml` - Test environment setup

## Running Tests

```bash
# Start test environment
docker-compose -f docker-compose.test.yml up -d

# Run all integration tests
go test ./tests/integration/... -v

# Run specific test suite
go test ./tests/integration/grpc -v
go test ./tests/integration/workflows -v

# Run with coverage
go test ./tests/integration/... -v -coverprofile=coverage.out

# Cleanup test environment
docker-compose -f docker-compose.test.yml down -v
```

## Test Environment

The integration tests use a dedicated test environment with:

- PostgreSQL test database
- MongoDB test instance
- Redis test instance
- RabbitMQ test broker
- All microservices in test mode

## Test Data

Test data is automatically seeded and cleaned up between test runs.
