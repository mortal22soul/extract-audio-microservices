# Tiltfile for Video Converter Microservices
load('ext://restart_process', 'docker_build_with_restart')

# Configure local Kubernetes cluster
allow_k8s_contexts(['docker-desktop', 'kind-kind', 'minikube'])

# Database services (deployed via plain Kubernetes manifests)
print("Setting up database services...")

k8s_yaml([
    'infrastructure/k8s/databases/mongodb.yaml',
    'infrastructure/k8s/databases/redis.yaml',
    'infrastructure/k8s/databases/rabbitmq.yaml',
    'infrastructure/k8s/storage/minio.yaml',
])

# Go services with live reload
print("Setting up Go services...")

# Gateway Service
docker_build_with_restart(
    'gateway-service',
    context='./services/gateway',
    dockerfile='./services/gateway/Dockerfile.dev',
    entrypoint=['./bin/gateway'],
    only=['./cmd', './internal', './go.mod', './go.sum'],
    live_update=[
        sync('./services/gateway', '/app'),
        run('go build -o bin/gateway ./cmd/main.go', trigger=['**/*.go'])
    ]
)

# Auth Service
docker_build_with_restart(
    'auth-service',
    context='./services/auth',
    dockerfile='./services/auth/Dockerfile.dev',
    entrypoint=['./bin/auth'],
    only=['./cmd', './internal', './go.mod', './go.sum'],
    live_update=[
        sync('./services/auth', '/app'),
        run('go build -o bin/auth ./cmd/main.go', trigger=['**/*.go'])
    ]
)

# Converter Service
docker_build_with_restart(
    'converter-service',
    context='./services/converter',
    dockerfile='./services/converter/Dockerfile.dev',
    entrypoint=['./bin/converter'],
    only=['./cmd', './internal', './go.mod', './go.sum'],
    live_update=[
        sync('./services/converter', '/app'),
        run('go build -o bin/converter ./cmd/main.go', trigger=['**/*.go'])
    ]
)

# Notification Service
docker_build_with_restart(
    'notification-service',
    context='./services/notification',
    dockerfile='./services/notification/Dockerfile.dev',
    entrypoint=['./bin/notification'],
    only=['./cmd', './internal', './go.mod', './go.sum'],
    live_update=[
        sync('./services/notification', '/app'),
        run('go build -o bin/notification ./cmd/main.go', trigger=['**/*.go'])
    ]
)

# TypeScript services with pnpm
print("Setting up TypeScript services...")

# Frontend Service (Next.js)
docker_build(
    'frontend-service',
    context='./services/frontend',
    dockerfile='./services/frontend/Dockerfile.dev',
    live_update=[
        sync('./services/frontend/src', '/app/src'),
        sync('./services/frontend/package.json', '/app/package.json'),
        sync('./services/frontend/pnpm-lock.yaml', '/app/pnpm-lock.yaml'),
        run('pnpm install', trigger=['package.json', 'pnpm-lock.yaml']),
    ]
)

# Realtime Service (Socket.IO)
docker_build(
    'realtime-service',
    context='./services/realtime',
    dockerfile='./services/realtime/Dockerfile.dev',
    live_update=[
        sync('./services/realtime/src', '/app/src'),
        sync('./services/realtime/package.json', '/app/package.json'),
        sync('./services/realtime/pnpm-lock.yaml', '/app/pnpm-lock.yaml'),
        run('pnpm install', trigger=['package.json', 'pnpm-lock.yaml']),
        run('pnpm run build', trigger=['src/**/*'])
    ]
)

# Python service with uv
print("Setting up Python service...")

# Analytics Service (FastAPI)
docker_build(
    'analytics-service',
    context='./services/analytics',
    dockerfile='./services/analytics/Dockerfile.dev',
    live_update=[
        sync('./services/analytics/src', '/app/src'),
        sync('./services/analytics/pyproject.toml', '/app/pyproject.toml'),
        run('uv sync', trigger=['pyproject.toml']),
        restart_container()
    ]
)

# Kubernetes deployments
print("Loading Kubernetes manifests...")

k8s_yaml([
    'infrastructure/k8s/configmaps/gateway-config.yaml',
    'infrastructure/k8s/secrets/app-secrets.yaml',
    'infrastructure/k8s/services/gateway.yaml',
    'infrastructure/k8s/services/auth.yaml',
    'infrastructure/k8s/services/converter.yaml',
    'infrastructure/k8s/services/notification.yaml',
    'infrastructure/k8s/services/analytics.yaml',
    'infrastructure/k8s/services/realtime.yaml',
    'infrastructure/k8s/services/frontend.yaml'
])

# Resource definitions and port forwards
print("Setting up resources and port forwarding...")

# Services
k8s_resource('gateway-service', port_forwards='8080:8080', resource_deps=['mongodb', 'redis'])
k8s_resource('frontend-service', port_forwards='3000:3000', resource_deps=['gateway-service', 'realtime-service'])
k8s_resource('realtime-service', port_forwards='3001:3001', resource_deps=['redis'])
k8s_resource('analytics-service', port_forwards='8000:8000', resource_deps=['mongodb', 'rabbitmq'])
k8s_resource('auth-service', resource_deps=['mongodb'])
k8s_resource('converter-service', resource_deps=['mongodb', 'rabbitmq'])
k8s_resource('notification-service', resource_deps=['rabbitmq'])

# Infrastructure
k8s_resource('mongodb', port_forwards='27017:27017')
k8s_resource('redis', port_forwards='6379:6379')
k8s_resource('rabbitmq', port_forwards=['5672:5672', '15672:15672'])
k8s_resource('minio', port_forwards=['9000:9000', '9001:9001'])

print("Tilt configuration loaded successfully!")
print("Run 'tilt up' to start all services")
print("Access Tilt dashboard at http://localhost:10350")