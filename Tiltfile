# Tiltfile for Video Converter Microservices
load('ext://helm_resource', 'helm_resource', 'helm_repo')
load('ext://restart_process', 'docker_build_with_restart')

# Configure local Kubernetes cluster
allow_k8s_contexts(['docker-desktop', 'kind-kind', 'minikube'])

# Add Bitnami Helm repository for databases
helm_repo('bitnami', 'https://charts.bitnami.com/bitnami')

# Database services
print("Setting up database services...")

helm_resource('postgresql',
              'bitnami/postgresql',
              flags=['--set', 'auth.postgresPassword=dev123',
                     '--set', 'auth.database=videoconverter',
                     '--set', 'primary.persistence.size=1Gi'],
              resource_deps=[])

helm_resource('mongodb',
              'bitnami/mongodb',
              flags=['--set', 'auth.rootPassword=dev123',
                     '--set', 'auth.database=videoconverter',
                     '--set', 'persistence.size=1Gi'],
              resource_deps=[])

helm_resource('redis',
              'bitnami/redis',
              flags=['--set', 'auth.password=dev123',
                     '--set', 'master.persistence.size=1Gi'],
              resource_deps=[])

helm_resource('rabbitmq',
              'bitnami/rabbitmq',
              flags=['--set', 'auth.username=admin',
                     '--set', 'auth.password=dev123',
                     '--set', 'persistence.size=1Gi'],
              resource_deps=[])

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
    'infrastructure/k8s/services/gateway.yaml',
    'infrastructure/k8s/services/auth.yaml',
    'infrastructure/k8s/services/converter.yaml',
    'infrastructure/k8s/services/notification.yaml',
    'infrastructure/k8s/services/analytics.yaml',
    'infrastructure/k8s/services/realtime.yaml',
    'infrastructure/k8s/services/frontend.yaml'
])

# Port forwards for local access
print("Setting up port forwarding...")

k8s_resource('gateway-service', port_forwards='8080:8080')
k8s_resource('frontend-service', port_forwards='3000:3000')
k8s_resource('realtime-service', port_forwards='3001:3001')
k8s_resource('analytics-service', port_forwards='8000:8000')

# Resource dependencies
k8s_resource('gateway-service', resource_deps=['postgresql', 'mongodb', 'redis'])
k8s_resource('auth-service', resource_deps=['postgresql'])
k8s_resource('converter-service', resource_deps=['mongodb', 'rabbitmq'])
k8s_resource('notification-service', resource_deps=['rabbitmq'])
k8s_resource('analytics-service', resource_deps=['mongodb', 'rabbitmq'])
k8s_resource('realtime-service', resource_deps=['redis'])

print("Tilt configuration loaded successfully!")
print("Run 'tilt up' to start all services")
print("Access Tilt dashboard at http://localhost:10350")