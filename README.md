# Video Converter Microservices

A modern, polyglot microservices architecture for video-to-MP3 conversion with real-time updates and
ML-powered analytics.

## Architecture

- **Gateway Service** (Go) - API gateway with authentication and file handling
- **Auth Service** (Go) - gRPC authentication service
- **Converter Service** (Go) - Video-to-MP3 conversion with FFmpeg
- **Notification Service** (Go) - Email notifications via SMTP
- **Analytics Service** (Python) - ML-powered video analysis with FastAPI
- **Realtime Service** (TypeScript) - WebSocket service for live updates
- **Frontend Service** (Next.js) - Modern web interface

## Quick Start

### Prerequisites

- Docker and Docker Compose
- Kubernetes (Docker Desktop, minikube, or kind)
- Tilt.dev (recommended for development)
- Node.js 18+ with pnpm
- Go 1.21+
- Python 3.11+ with uv

### Option 1: Tilt.dev (Recommended)

```bash
# Install dependencies
make deps

# Start all services with live reload
make dev
# or
tilt up

# Access Tilt dashboard
open http://localhost:10350
```

### Option 2: Docker Compose

```bash
# Copy environment file
cp .env.example .env

# Start all services
make docker-up
# or
docker-compose up -d

# View logs
docker-compose logs -f
```

## Service Endpoints

- **Frontend**: <http://localhost:3000>
- **Gateway API**: <http://localhost:8080>
- **Analytics API**: <http://localhost:8000>
- **Realtime WebSocket**: <http://localhost:3001>
- **RabbitMQ Management**: <http://localhost:15672> (admin/dev123)
- **MinIO Console**: <http://localhost:9001> (admin/dev123456)

## Development

### Building Services

```bash
# Build all services
make build

# Build specific service
cd services/gateway && go build -o bin/gateway ./cmd/main.go
```

### Running Tests

```bash
# Run all tests
make test

# Run specific service tests
cd services/gateway && go test ./...
```

### Code Generation

```bash
# Generate protobuf code
make proto
# or
buf generate
```

### Linting and Formatting

```bash
# Lint all services
make lint

# Format all code
make format
```

## Project Structure

```txt
├── services/                 # Microservices
│   ├── gateway/             # Go API gateway
│   ├── auth/                # Go authentication service
│   ├── converter/           # Go video conversion service
│   ├── notification/        # Go notification service
│   ├── analytics/           # Python ML service
│   ├── realtime/            # TypeScript WebSocket service
│   └── frontend/            # Next.js web application
├── shared/                  # Shared code and definitions
│   └── proto/               # Protocol Buffer definitions
├── infrastructure/          # Infrastructure as code
│   ├── k8s/                 # Kubernetes manifests
│   └── docker/              # Docker configurations
├── Tiltfile                 # Tilt.dev configuration
├── docker-compose.yml       # Docker Compose configuration
└── Makefile                 # Development commands
```

## Technology Stack

### Backend Services

- **Go 1.21+** - High-performance backend services
- **gRPC** - Inter-service communication
- **Data Stores**:
  - **MongoDB**: Users, sessions, video metadata, and conversion job status
  - **Redis**: Session caching, rate limiting, and real-time messaging
  - **RabbitMQ**: Asynchronous task queues for video processing
  - **MinIO**: Object storage for original and converted video files
- **FFmpeg** - Video processing

### Frontend & Real-time

- **Next.js 14** - React framework with App Router
- **TypeScript** - Type-safe JavaScript
- **Socket.IO** - WebSocket communication
- **Tailwind CSS** - Utility-first CSS

### ML & Analytics

- **Python 3.11+** - ML and data processing
- **FastAPI** - High-performance Python API
- **OpenCV** - Computer vision
- **scikit-learn** - Machine learning
- **Transformers** - NLP models

### Development Tools

- **Tilt.dev** - Local development environment
- **pnpm** - Fast Node.js package manager
- **uv** - Fast Python package manager
- **buf** - Protocol Buffer toolchain
- **Docker** - Containerization
- **Kubernetes** - Container orchestration

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests and linting
5. Submit a pull request

## License

MIT License - see LICENSE file for details
