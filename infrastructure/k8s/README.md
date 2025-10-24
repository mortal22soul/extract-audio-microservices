# Kubernetes Manifests

This directory contains Kubernetes deployment manifests for all services.

## Structure

- `databases/` - Database StatefulSets and Services
- `services/` - Application Deployments and Services
- `ingress/` - Ingress controllers and routing
- `monitoring/` - Prometheus, Grafana, and observability tools

## Deployment

Use Helm charts in the `helm/` directory for production deployments.
For local development, use Tilt.dev configuration in the root directory.