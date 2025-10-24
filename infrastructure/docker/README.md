# Docker Configuration

This directory contains Docker-related configuration files.

## Files

- `docker-compose.yml` - Local development environment
- `docker-compose.prod.yml` - Production-like environment
- `Dockerfiles/` - Service-specific Dockerfiles

## Usage

For local development without Kubernetes:
```bash
docker-compose up -d
```

For production-like testing:
```bash
docker-compose -f docker-compose.prod.yml up -d
```