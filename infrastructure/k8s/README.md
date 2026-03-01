# Kubernetes Manifests

This directory contains Kubernetes deployment manifests for all services.

## Structure

- `databases/` - Database StatefulSets and Services
- `services/` - Application Deployments and Services
- `ingress/` - Ingress controllers and routing
- `monitoring/` - Prometheus, Grafana, and observability tools

## Deployment

Apply manifests directly with kubectl, or use Tilt.dev for local development (see root `Tiltfile`).
