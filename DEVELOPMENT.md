# Development Guide

This guide covers the development setup and workflow for the Video Converter microservices project.

## Prerequisites

### Required Tools

- **Go 1.21+** - Backend services
- **Node.js 18+** - Frontend and realtime services
- **pnpm 8+** - Node.js package manager
- **Python 3.11+** - Analytics service
- **uv 0.4+** - Python package manager
- **Docker & Docker Compose** - Container runtime
- **Make** - Build automation

### Optional Tools (for Kubernetes development)

- **kubectl** - Kubernetes CLI
- **Tilt** - Local Kubernetes development

### Development Tools

- **golangci-lint** - Go linter
- **buf** - Protocol buffer tooling
- **pre-commit** - Git hooks

## Quick Start

### Automated Setup

Run the setup script to install dependencies and configure the development environment:

**Linux/macOS:**

```bash
./scripts/setup-dev.sh
```

**Windows:**

```cmd
scripts\setup-dev.bat
```

### Manual Setup

1. **Install dependencies:**

   ```bash
   make deps
   ```

2. **Install development tools:**

   ```bash
   # Go tools
   go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
   go install golang.org/x/tools/cmd/goimports@latest

   # Pre-commit hooks
   pip install pre-commit
   make pre-commit-install
   ```

3. **Generate protocol buffers:**
   ```bash
   make proto
   ```

## Development Workflows

### Option 1: Tilt.dev (Recommended for Kubernetes)

Start all services with live reload in Kubernetes:

```bash
make dev
# or
tilt up
```

Access the Tilt dashboard at http://localhost:10350

### Option 2: Docker Compose (Fallback)

Start all services with Docker Compose:

```bash
make docker-up
```

Stop services:

```bash
make docker-down
```

### Option 3: Individual Services

Start services individually for focused development:

```bash
# Start databases only
docker-compose up -d postgresql mongodb redis rabbitmq

# Run individual services
cd services/auth && go run cmd/main.go
cd services/gateway && go run cmd/main.go
cd services/frontend && pnpm dev
cd services/realtime && pnpm dev
cd services/analytics && uv run uvicorn src.main:app --reload
```

## Code Quality

### Linting

Run linters for all services:

```bash
make lint
```

Run linters for specific languages:

```bash
make lint-go      # Go services
make lint-ts      # TypeScript services
make lint-python  # Python service
```

Auto-fix linting issues:

```bash
make lint-fix
```

### Formatting

Format all code:

```bash
make format
```

Format specific languages:

```bash
make format-go      # Go services
make format-ts      # TypeScript services
make format-python  # Python service
```

### Pre-commit Hooks

Pre-commit hooks automatically run linting and formatting on changed files:

```bash
# Install hooks (done automatically by setup script)
make pre-commit-install

# Run hooks manually on all files
make pre-commit-run

# Update hooks to latest versions
make pre-commit-update
```

### Type Checking

Run TypeScript type checking:

```bash
pnpm type-check
```

Run Python type checking:

```bash
cd services/analytics && uv run mypy src/
```

## Testing

### Run All Tests

```bash
make test
```

### Run Tests by Language

```bash
make check-go      # Go tests + linting
make check-ts      # TypeScript tests + linting
make check-python  # Python tests + linting
```

### Individual Service Tests

```bash
# Go services
cd services/auth && go test ./...
cd services/gateway && go test ./...

# TypeScript services
cd services/frontend && pnpm test
cd services/realtime && pnpm test

# Python service
cd services/analytics && uv run pytest
```

## Building

### Build All Services

```bash
make build
```

### Build Individual Services

```bash
# Go services
cd services/auth && go build -o bin/auth ./cmd/main.go

# TypeScript services
cd services/frontend && pnpm build
cd services/realtime && pnpm build

# Python service (no build step needed)
cd services/analytics && uv sync
```

## Protocol Buffers

Generate code from .proto files:

```bash
make proto
```

Proto files are located in `shared/proto/` and generated code goes to each service's appropriate
directory.

## Environment Variables

### Development Environment

Copy the example environment file:

```bash
cp .env.example .env
```

### Service-Specific Variables

Each service has its own environment configuration:

- **Auth Service**: Database connection, JWT secrets
- **Gateway Service**: Service URLs, file storage
- **Converter Service**: Processing settings, queue config
- **Analytics Service**: ML model settings, API keys
- **Realtime Service**: WebSocket settings, Redis config
- **Frontend Service**: API endpoints, feature flags

## Debugging

### Service Logs

**With Tilt:**

- View logs in the Tilt dashboard
- Or use: `tilt logs <service-name>`

**With Docker Compose:**

```bash
docker-compose logs -f <service-name>
```

### Database Access

**PostgreSQL:**

```bash
docker-compose exec postgresql psql -U postgres -d videoconverter
```

**MongoDB:**

```bash
docker-compose exec mongodb mongosh -u admin -p dev123
```

**Redis:**

```bash
docker-compose exec redis redis-cli -a dev123
```

### Message Queue

**RabbitMQ Management UI:** http://localhost:15672 (admin/dev123)

## Performance Profiling

### Go Services

Enable pprof endpoints in development:

```go
import _ "net/http/pprof"
```

Access profiling at: http://localhost:6060/debug/pprof/

### Node.js Services

Use built-in profiler:

```bash
node --inspect dist/index.js
```

## Common Issues

### Port Conflicts

If ports are already in use, update the port mappings in:

- `docker-compose.yml`
- `Tiltfile`
- Service configuration files

### Permission Issues

**Linux/macOS:**

```bash
sudo chown -R $USER:$USER .
```

**Windows:** Run terminal as Administrator if needed.

### Database Connection Issues

1. Ensure databases are running: `docker-compose ps`
2. Check connection strings in environment variables
3. Verify database initialization scripts

### Build Failures

1. Clean build artifacts: `make clean`
2. Reinstall dependencies: `make deps`
3. Check tool versions match requirements

## IDE Configuration

### VS Code

Recommended extensions:

- Go
- TypeScript and JavaScript Language Features
- Python
- ESLint
- Prettier
- Docker
- Kubernetes

### Settings

```json
{
  "go.lintTool": "golangci-lint",
  "go.formatTool": "goimports",
  "editor.formatOnSave": true,
  "editor.codeActionsOnSave": {
    "source.fixAll.eslint": true
  },
  "python.defaultInterpreterPath": "./services/analytics/.venv/bin/python"
}
```

## Contributing

1. **Create a feature branch:**

   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make changes and test:**

   ```bash
   make check
   ```

3. **Commit with pre-commit hooks:**

   ```bash
   git add .
   git commit -m "feat: your feature description"
   ```

4. **Push and create PR:**
   ```bash
   git push origin feature/your-feature-name
   ```

## Makefile Commands

Run `make help` to see all available commands:

```bash
make help
```

Key commands:

- `make dev` - Start development environment
- `make test` - Run all tests
- `make lint` - Run all linters
- `make format` - Format all code
- `make build` - Build all services
- `make clean` - Clean build artifacts
- `make setup-dev` - Set up development environment
