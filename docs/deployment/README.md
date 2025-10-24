# Deployment Guide

This guide provides comprehensive instructions for deploying the Video Converter microservices
platform across different environments.

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Local Development](#local-development)
3. [Staging Deployment](#staging-deployment)
4. [Production Deployment](#production-deployment)
5. [Configuration Management](#configuration-management)
6. [Monitoring and Observability](#monitoring-and-observability)
7. [Troubleshooting](#troubleshooting)

## Prerequisites

### Required Tools

- **Docker**: Version 20.10+
- **Kubernetes**: Version 1.25+
- **Helm**: Version 3.10+
- **kubectl**: Compatible with your Kubernetes version
- **Tilt**: Version 0.30+ (for local development)

### Development Tools

- **Go**: Version 1.21+
- **Node.js**: Version 18+
- **Python**: Version 3.11+
- **pnpm**: Version 8.0+
- **uv**: Version 0.4+ (Python package manager)

### Infrastructure Requirements

#### Minimum Resources (Development)

- **CPU**: 4 cores
- **Memory**: 8GB RAM
- **Storage**: 50GB available space
- **Network**: Stable internet connection

#### Production Resources

- **CPU**: 16+ cores (distributed across nodes)
- **Memory**: 32GB+ RAM (distributed across nodes)
- **Storage**: 500GB+ SSD storage
- **Network**: High-bandwidth, low-latency network

## Local Development

### Option 1: Tilt.dev (Recommended)

Tilt provides the fastest development experience with live reload and unified service management.

#### Setup

1. **Install Tilt**:

   ```bash
   # macOS
   brew install tilt-dev/tap/tilt

   # Linux
   curl -fsSL https://raw.githubusercontent.com/tilt-dev/tilt/master/scripts/install.sh | bash

   # Windows
   iex ((new-object net.webclient).downloadstring('https://raw.githubusercontent.com/tilt-dev/tilt/master/scripts/install.ps1'))
   ```

2. **Start Local Kubernetes**:

   ```bash
   # Docker Desktop (recommended)
   # Enable Kubernetes in Docker Desktop settings

   # Or use kind
   kind create cluster --name video-converter

   # Or use minikube
   minikube start --memory=8192 --cpus=4
   ```

3. **Start Development Environment**:

   ```bash
   # Clone repository
   git clone <repository-url>
   cd video-converter

   # Start all services
   tilt up

   # Open Tilt dashboard
   # Navigate to http://localhost:10350
   ```

4. **Access Services**:
   - **Frontend**: http://localhost:3000
   - **API Gateway**: http://localhost:8080
   - **Realtime Service**: http://localhost:3001
   - **Analytics API**: http://localhost:8000
   - **Tilt Dashboard**: http://localhost:10350

#### Development Workflow

```bash
# View logs for specific service
tilt logs gateway-service

# Restart specific service
tilt trigger auth-service

# Stop all services
tilt down

# Force rebuild
tilt trigger --build gateway-service
```

### Option 2: Docker Compose

For developers who prefer Docker Compose or don't have Kubernetes available.

#### Setup

1. **Start Infrastructure Services**:

   ```bash
   # Start databases and message brokers
   docker-compose -f docker-compose.yml up -d postgres mongodb redis rabbitmq

   # Wait for services to be ready
   docker-compose logs -f postgres
   ```

2. **Build and Start Application Services**:

   ```bash
   # Build all services
   make build-all

   # Start application services
   docker-compose up -d gateway auth converter analytics realtime frontend notification
   ```

3. **Initialize Databases**:

   ```bash
   # Run database migrations
   make migrate-up

   # Seed development data
   make seed-dev
   ```

#### Development Commands

```bash
# View logs
docker-compose logs -f gateway

# Restart service
docker-compose restart auth

# Rebuild and restart
docker-compose up -d --build gateway

# Stop all services
docker-compose down

# Clean up volumes
docker-compose down -v
```

### Environment Variables

Create a `.env` file in the project root:

```bash
# Database Configuration
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_DB=video_converter
POSTGRES_USER=postgres
POSTGRES_PASSWORD=dev123

MONGODB_URI=mongodb://localhost:27017/video_converter
REDIS_URL=redis://localhost:6379
RABBITMQ_URL=amqp://admin:dev123@localhost:5672/

# JWT Configuration
JWT_SECRET=your-super-secret-jwt-key-change-in-production
JWT_EXPIRY=24h
REFRESH_TOKEN_EXPIRY=168h

# Service URLs
AUTH_SERVICE_URL=auth-service:50051
ANALYTICS_SERVICE_URL=analytics-service:50052
REALTIME_SERVICE_URL=http://realtime-service:3001

# File Storage
UPLOAD_MAX_SIZE=100MB
STORAGE_PATH=/tmp/uploads

# Email Configuration (for notifications)
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=your-email@gmail.com
SMTP_PASSWORD=your-app-password
```

## Staging Deployment

### Prerequisites

- Kubernetes cluster (EKS, GKE, AKS, or self-managed)
- Helm 3.10+
- kubectl configured for your cluster
- Container registry (ECR, GCR, ACR, or Docker Hub)

### Setup

1. **Create Namespace**:

   ```bash
   kubectl create namespace video-converter-staging
   kubectl config set-context --current --namespace=video-converter-staging
   ```

2. **Install Dependencies**:

   ```bash
   # Add Helm repositories
   helm repo add bitnami https://charts.bitnami.com/bitnami
   helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
   helm repo add grafana https://grafana.github.io/helm-charts
   helm repo update

   # Install databases
   helm install postgresql bitnami/postgresql \
     --set auth.postgresPassword=staging123 \
     --set primary.persistence.size=20Gi

   helm install mongodb bitnami/mongodb \
     --set auth.rootPassword=staging123 \
     --set persistence.size=20Gi

   helm install redis bitnami/redis \
     --set auth.password=staging123

   helm install rabbitmq bitnami/rabbitmq \
     --set auth.username=admin \
     --set auth.password=staging123
   ```

3. **Build and Push Images**:

   ```bash
   # Set your container registry
   export REGISTRY=your-registry.com/video-converter

   # Build and push all images
   make build-push-all REGISTRY=$REGISTRY TAG=staging
   ```

4. **Deploy Application**:

   ```bash
   # Install application using Helm
   helm install video-converter ./infrastructure/helm/video-converter \
     --values ./infrastructure/helm/video-converter/values-staging.yaml \
     --set image.registry=$REGISTRY \
     --set image.tag=staging
   ```

5. **Configure Ingress**:

   ```bash
   # Install ingress controller (if not already installed)
   helm install ingress-nginx ingress-nginx/ingress-nginx

   # Apply ingress configuration
   kubectl apply -f infrastructure/k8s/ingress-staging.yaml
   ```

### Verification

```bash
# Check pod status
kubectl get pods

# Check services
kubectl get services

# Check ingress
kubectl get ingress

# View logs
kubectl logs -l app=gateway-service

# Port forward for testing
kubectl port-forward svc/gateway-service 8080:8080
```

## Production Deployment

### Infrastructure Setup

#### AWS EKS Example

1. **Create EKS Cluster**:

   ```bash
   # Using eksctl
   eksctl create cluster \
     --name video-converter-prod \
     --version 1.25 \
     --region us-west-2 \
     --nodegroup-name standard-workers \
     --node-type m5.xlarge \
     --nodes 3 \
     --nodes-min 1 \
     --nodes-max 10 \
     --managed
   ```

2. **Configure Storage Classes**:

   ```bash
   # Apply storage class for databases
   kubectl apply -f infrastructure/k8s/storage/aws-ebs-storage-class.yaml
   ```

3. **Install Cluster Autoscaler**:
   ```bash
   helm install cluster-autoscaler autoscaler/cluster-autoscaler \
     --set autoDiscovery.clusterName=video-converter-prod \
     --set awsRegion=us-west-2
   ```

#### Database Setup (Production)

For production, use managed database services:

1. **Amazon RDS (PostgreSQL)**:

   ```bash
   # Create RDS instance
   aws rds create-db-instance \
     --db-instance-identifier video-converter-prod \
     --db-instance-class db.r5.large \
     --engine postgres \
     --engine-version 14.9 \
     --allocated-storage 100 \
     --storage-type gp2 \
     --storage-encrypted \
     --master-username postgres \
     --master-user-password <secure-password> \
     --vpc-security-group-ids sg-xxxxxxxxx \
     --db-subnet-group-name video-converter-subnet-group \
     --backup-retention-period 7 \
     --multi-az
   ```

2. **Amazon DocumentDB (MongoDB)**:

   ```bash
   # Create DocumentDB cluster
   aws docdb create-db-cluster \
     --db-cluster-identifier video-converter-prod \
     --engine docdb \
     --master-username admin \
     --master-user-password <secure-password> \
     --vpc-security-group-ids sg-xxxxxxxxx \
     --db-subnet-group-name video-converter-subnet-group \
     --storage-encrypted \
     --backup-retention-period 7
   ```

3. **Amazon ElastiCache (Redis)**:
   ```bash
   # Create Redis cluster
   aws elasticache create-replication-group \
     --replication-group-id video-converter-prod \
     --description "Video Converter Redis Cluster" \
     --cache-node-type cache.r5.large \
     --engine redis \
     --num-cache-clusters 3 \
     --security-group-ids sg-xxxxxxxxx \
     --subnet-group-name video-converter-cache-subnet-group \
     --at-rest-encryption-enabled \
     --transit-encryption-enabled
   ```

### Production Deployment Steps

1. **Create Production Namespace**:

   ```bash
   kubectl create namespace video-converter-prod
   kubectl config set-context --current --namespace=video-converter-prod
   ```

2. **Create Secrets**:

   ```bash
   # Database secrets
   kubectl create secret generic database-secrets \
     --from-literal=postgres-password=<secure-password> \
     --from-literal=mongodb-password=<secure-password> \
     --from-literal=redis-password=<secure-password>

   # Application secrets
   kubectl create secret generic app-secrets \
     --from-literal=jwt-secret=<secure-jwt-secret> \
     --from-literal=smtp-password=<smtp-password>

   # TLS certificates
   kubectl create secret tls video-converter-tls \
     --cert=path/to/tls.crt \
     --key=path/to/tls.key
   ```

3. **Deploy Monitoring Stack**:

   ```bash
   # Install Prometheus
   helm install prometheus prometheus-community/kube-prometheus-stack \
     --values infrastructure/helm/monitoring/prometheus-values.yaml

   # Install Grafana
   helm install grafana grafana/grafana \
     --values infrastructure/helm/monitoring/grafana-values.yaml

   # Install Jaeger
   kubectl apply -f infrastructure/k8s/monitoring/jaeger.yaml
   ```

4. **Deploy Application**:

   ```bash
   # Build and push production images
   export REGISTRY=your-prod-registry.com/video-converter
   make build-push-all REGISTRY=$REGISTRY TAG=v1.0.0

   # Deploy using Helm
   helm install video-converter ./infrastructure/helm/video-converter \
     --values ./infrastructure/helm/video-converter/values-prod.yaml \
     --set image.registry=$REGISTRY \
     --set image.tag=v1.0.0
   ```

5. **Configure Load Balancer**:
   ```bash
   # Apply production ingress with SSL
   kubectl apply -f infrastructure/k8s/ingress-prod.yaml
   ```

### Production Configuration

#### Horizontal Pod Autoscaler

```yaml
# infrastructure/k8s/hpa/gateway-hpa.yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: gateway-service-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: gateway-service
  minReplicas: 3
  maxReplicas: 20
  metrics:
    - type: Resource
      resource:
        name: cpu
        target:
          type: Utilization
          averageUtilization: 70
    - type: Resource
      resource:
        name: memory
        target:
          type: Utilization
          averageUtilization: 80
```

#### Pod Disruption Budget

```yaml
# infrastructure/k8s/pdb/gateway-pdb.yaml
apiVersion: policy/v1
kind: PodDisruptionBudget
metadata:
  name: gateway-service-pdb
spec:
  minAvailable: 2
  selector:
    matchLabels:
      app: gateway-service
```

## Configuration Management

### Environment-Specific Values

#### Development (values-dev.yaml)

```yaml
replicaCount: 1
image:
  pullPolicy: Always
resources:
  requests:
    memory: '128Mi'
    cpu: '100m'
  limits:
    memory: '512Mi'
    cpu: '500m'
autoscaling:
  enabled: false
```

#### Staging (values-staging.yaml)

```yaml
replicaCount: 2
image:
  pullPolicy: Always
resources:
  requests:
    memory: '256Mi'
    cpu: '200m'
  limits:
    memory: '1Gi'
    cpu: '1000m'
autoscaling:
  enabled: true
  minReplicas: 2
  maxReplicas: 5
```

#### Production (values-prod.yaml)

```yaml
replicaCount: 3
image:
  pullPolicy: IfNotPresent
resources:
  requests:
    memory: '512Mi'
    cpu: '500m'
  limits:
    memory: '2Gi'
    cpu: '2000m'
autoscaling:
  enabled: true
  minReplicas: 3
  maxReplicas: 20
```

### ConfigMaps

```yaml
# infrastructure/k8s/configmaps/gateway-config.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: gateway-config
data:
  AUTH_SERVICE_URL: 'auth-service:50051'
  ANALYTICS_SERVICE_URL: 'analytics-service:50052'
  UPLOAD_MAX_SIZE: '100MB'
  CORS_ORIGINS: 'https://app.videoconverter.com'
  LOG_LEVEL: 'info'
```

## Monitoring and Observability

### Metrics Collection

The system exposes Prometheus metrics on `/metrics` endpoints:

- **Gateway Service**: `:8080/metrics`
- **Auth Service**: `:50051/metrics`
- **Converter Service**: `:50052/metrics`
- **Analytics Service**: `:8000/metrics`
- **Realtime Service**: `:3001/metrics`
- **Notification Service**: `:50053/metrics`

### Key Metrics to Monitor

#### Application Metrics

- Request rate and latency
- Error rates by service
- Video conversion queue length
- Active WebSocket connections
- Database connection pool usage

#### Infrastructure Metrics

- CPU and memory usage
- Disk I/O and storage usage
- Network traffic
- Pod restart counts
- Node resource utilization

### Alerting Rules

```yaml
# infrastructure/k8s/monitoring/alerts.yaml
groups:
  - name: video-converter.rules
    rules:
      - alert: HighErrorRate
        expr: rate(http_requests_total{status=~"5.."}[5m]) > 0.1
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: 'High error rate detected'
          description: 'Error rate is {{ $value }} errors per second'

      - alert: HighMemoryUsage
        expr: container_memory_usage_bytes / container_spec_memory_limit_bytes > 0.9
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: 'High memory usage'
          description: 'Memory usage is above 90%'
```

### Grafana Dashboards

Import the provided dashboards:

- **System Overview**: `infrastructure/k8s/monitoring/dashboards/system-overview.json`
- **Service Metrics**: `infrastructure/k8s/monitoring/dashboards/service-metrics.json`
- **Database Metrics**: `infrastructure/k8s/monitoring/dashboards/database-metrics.json`

## Troubleshooting

### Common Issues

#### 1. Pod Startup Failures

```bash
# Check pod status
kubectl get pods

# Describe pod for events
kubectl describe pod <pod-name>

# Check logs
kubectl logs <pod-name> --previous

# Check resource constraints
kubectl top pods
```

#### 2. Service Discovery Issues

```bash
# Check service endpoints
kubectl get endpoints

# Test service connectivity
kubectl run debug --image=busybox -it --rm -- nslookup auth-service

# Check network policies
kubectl get networkpolicies
```

#### 3. Database Connection Issues

```bash
# Check database pod status
kubectl get pods -l app=postgresql

# Test database connectivity
kubectl run postgres-client --image=postgres:14 -it --rm -- psql -h postgresql -U postgres

# Check secrets
kubectl get secrets database-secrets -o yaml
```

#### 4. Performance Issues

```bash
# Check resource usage
kubectl top pods
kubectl top nodes

# Check HPA status
kubectl get hpa

# Check metrics
kubectl port-forward svc/prometheus-server 9090:80
# Navigate to http://localhost:9090
```

### Log Analysis

#### Centralized Logging with ELK Stack

```bash
# Deploy Elasticsearch
kubectl apply -f infrastructure/k8s/monitoring/elasticsearch.yaml

# Deploy Logstash
kubectl apply -f infrastructure/k8s/monitoring/logstash.yaml

# Deploy Filebeat
kubectl apply -f infrastructure/k8s/monitoring/filebeat.yaml

# Access Kibana
kubectl port-forward svc/kibana 5601:5601
```

#### Log Queries

Common log queries for troubleshooting:

```bash
# Error logs from all services
level:ERROR

# Slow requests (>1s)
duration:>1000 AND path:/api/*

# Authentication failures
message:"authentication failed" OR message:"invalid token"

# Video conversion errors
service:converter AND level:ERROR
```

### Health Checks

All services expose health check endpoints:

```bash
# Check service health
curl http://gateway-service:8080/health
curl http://auth-service:50051/health
curl http://analytics-service:8000/health
```

### Backup and Recovery

#### Database Backups

```bash
# PostgreSQL backup
kubectl exec -it postgresql-0 -- pg_dump -U postgres video_converter > backup.sql

# MongoDB backup
kubectl exec -it mongodb-0 -- mongodump --uri="mongodb://admin:password@localhost:27017/video_converter"

# Restore PostgreSQL
kubectl exec -i postgresql-0 -- psql -U postgres video_converter < backup.sql
```

#### Automated Backup Jobs

```yaml
# infrastructure/k8s/databases/backup-cronjobs.yaml
apiVersion: batch/v1
kind: CronJob
metadata:
  name: postgres-backup
spec:
  schedule: '0 2 * * *' # Daily at 2 AM
  jobTemplate:
    spec:
      template:
        spec:
          containers:
            - name: postgres-backup
              image: postgres:14
              command:
                - /bin/bash
                - -c
                - |
                  pg_dump -h postgresql -U postgres video_converter | \
                  gzip > /backup/postgres-$(date +%Y%m%d).sql.gz
              volumeMounts:
                - name: backup-storage
                  mountPath: /backup
          restartPolicy: OnFailure
          volumes:
            - name: backup-storage
              persistentVolumeClaim:
                claimName: backup-pvc
```

### Disaster Recovery

#### RTO/RPO Targets

- **Recovery Time Objective (RTO)**: 4 hours
- **Recovery Point Objective (RPO)**: 1 hour

#### Recovery Procedures

1. **Service Recovery**:

   ```bash
   # Restore from backup
   helm rollback video-converter <revision>

   # Scale up services
   kubectl scale deployment gateway-service --replicas=3
   ```

2. **Database Recovery**:

   ```bash
   # Restore from latest backup
   kubectl exec -i postgresql-0 -- psql -U postgres video_converter < latest-backup.sql
   ```

3. **Verification**:

   ```bash
   # Run health checks
   make health-check

   # Verify functionality
   make integration-test
   ```

For additional support, refer to the [troubleshooting guide](troubleshooting.md) or contact the
development team.
