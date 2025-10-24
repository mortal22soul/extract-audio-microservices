# Configuration Guide

This guide provides comprehensive information about configuring the Video Converter microservices
platform for different environments and use cases.

## Table of Contents

1. [Configuration Overview](#configuration-overview)
2. [Environment Variables](#environment-variables)
3. [ConfigMaps and Secrets](#configmaps-and-secrets)
4. [Service Configuration](#service-configuration)
5. [Database Configuration](#database-configuration)
6. [Security Configuration](#security-configuration)
7. [Performance Tuning](#performance-tuning)
8. [Monitoring Configuration](#monitoring-configuration)

## Configuration Overview

The platform uses a hierarchical configuration system:

1. **Default Values**: Built into application code
2. **ConfigMaps**: Non-sensitive configuration data
3. **Secrets**: Sensitive configuration data (passwords, keys)
4. **Environment Variables**: Runtime configuration
5. **Command Line Arguments**: Override specific settings

### Configuration Precedence

```
Command Line Args > Environment Variables > Secrets > ConfigMaps > Default Values
```

## Environment Variables

### Common Environment Variables

All services support these common environment variables:

```bash
# Logging Configuration
LOG_LEVEL=info                    # debug, info, warn, error
LOG_FORMAT=json                   # json, text
LOG_OUTPUT=stdout                 # stdout, file

# Health Check Configuration
HEALTH_CHECK_PORT=8080
HEALTH_CHECK_PATH=/health
READINESS_CHECK_PATH=/ready

# Metrics Configuration
METRICS_PORT=8080
METRICS_PATH=/metrics
METRICS_ENABLED=true

# Tracing Configuration
JAEGER_ENDPOINT=http://jaeger:14268/api/traces
TRACING_ENABLED=true
TRACE_SAMPLE_RATE=0.1

# Service Discovery
SERVICE_NAME=gateway-service
SERVICE_VERSION=1.0.0
NAMESPACE=video-converter
```

### Service-Specific Environment Variables

#### Gateway Service

```bash
# Server Configuration
PORT=8080
HOST=0.0.0.0
CORS_ORIGINS=https://app.videoconverter.com,http://localhost:3000
CORS_METHODS=GET,POST,PUT,DELETE,OPTIONS
CORS_HEADERS=Content-Type,Authorization

# File Upload Configuration
UPLOAD_MAX_SIZE=100MB
UPLOAD_ALLOWED_TYPES=video/mp4,video/avi,video/mov,video/mkv
UPLOAD_TEMP_DIR=/tmp/uploads
UPLOAD_CLEANUP_INTERVAL=1h

# Rate Limiting
RATE_LIMIT_ENABLED=true
RATE_LIMIT_REQUESTS_PER_MINUTE=60
RATE_LIMIT_BURST=10

# gRPC Client Configuration
AUTH_SERVICE_URL=auth-service:50051
ANALYTICS_SERVICE_URL=analytics-service:50052
GRPC_TIMEOUT=30s
GRPC_MAX_RETRY_ATTEMPTS=3

# Database Configuration
MONGODB_URI=mongodb://mongodb:27017/video_converter
MONGODB_MAX_POOL_SIZE=100
MONGODB_TIMEOUT=30s

# Redis Configuration
REDIS_URL=redis://redis:6379
REDIS_MAX_CONNECTIONS=100
REDIS_TIMEOUT=5s
```

#### Auth Service

```bash
# gRPC Server Configuration
GRPC_PORT=50051
GRPC_HOST=0.0.0.0
GRPC_MAX_CONNECTIONS=1000

# Database Configuration
POSTGRES_HOST=postgresql
POSTGRES_PORT=5432
POSTGRES_DB=video_converter
POSTGRES_USER=postgres
POSTGRES_PASSWORD_FILE=/etc/secrets/postgres-password
POSTGRES_SSL_MODE=require
POSTGRES_MAX_CONNECTIONS=50
POSTGRES_MAX_IDLE_CONNECTIONS=10
POSTGRES_CONNECTION_TIMEOUT=30s

# JWT Configuration
JWT_SECRET_FILE=/etc/secrets/jwt-secret
JWT_EXPIRY=24h
JWT_REFRESH_EXPIRY=168h
JWT_ISSUER=video-converter-auth
JWT_AUDIENCE=video-converter-api

# Password Security
BCRYPT_COST=12
PASSWORD_MIN_LENGTH=8
PASSWORD_REQUIRE_UPPERCASE=true
PASSWORD_REQUIRE_LOWERCASE=true
PASSWORD_REQUIRE_NUMBERS=true
PASSWORD_REQUIRE_SYMBOLS=true

# Session Management
SESSION_TIMEOUT=24h
MAX_SESSIONS_PER_USER=5
SESSION_CLEANUP_INTERVAL=1h
```

#### Converter Service

```bash
# Worker Configuration
MAX_CONCURRENT_JOBS=4
WORKER_TIMEOUT=30m
JOB_RETRY_ATTEMPTS=3
JOB_RETRY_DELAY=5m

# FFmpeg Configuration
FFMPEG_PATH=/usr/bin/ffmpeg
FFMPEG_THREADS=0                  # 0 = auto-detect
FFMPEG_PRESET=medium              # ultrafast, superfast, veryfast, faster, fast, medium, slow, slower, veryslow
FFMPEG_QUALITY=192k               # Audio bitrate

# File Processing
TEMP_DIR=/tmp/conversion
CLEANUP_TEMP_FILES=true
MAX_FILE_SIZE=1GB

# RabbitMQ Configuration
RABBITMQ_URL=amqp://admin:password@rabbitmq:5672/
RABBITMQ_QUEUE=video_conversion
RABBITMQ_EXCHANGE=video_events
RABBITMQ_ROUTING_KEY=conversion.request
RABBITMQ_PREFETCH_COUNT=1
RABBITMQ_AUTO_ACK=false

# MongoDB Configuration (for file storage)
MONGODB_URI=mongodb://mongodb:27017/video_converter
GRIDFS_BUCKET=videos
GRIDFS_CHUNK_SIZE=261120          # 255KB chunks

# Progress Reporting
PROGRESS_UPDATE_INTERVAL=5s
REDIS_PROGRESS_KEY_PREFIX=conversion:progress:
```

#### Analytics Service

```bash
# FastAPI Configuration
HOST=0.0.0.0
PORT=8000
WORKERS=4
RELOAD=false

# ML Model Configuration
MODEL_PATH=/app/models
MODEL_CACHE_SIZE=3
MODEL_LOAD_TIMEOUT=60s
GPU_ENABLED=false
GPU_MEMORY_FRACTION=0.8

# Video Processing
MAX_VIDEO_DURATION=3600           # 1 hour max
THUMBNAIL_COUNT=5
THUMBNAIL_SIZE=320x240
THUMBNAIL_QUALITY=85

# Content Analysis
CONTENT_SAFETY_THRESHOLD=0.8
QUALITY_ANALYSIS_ENABLED=true
RECOMMENDATION_CACHE_TTL=3600

# Database Configuration
MONGODB_URI=mongodb://mongodb:27017/video_converter
MONGODB_COLLECTION_VIDEOS=videos
MONGODB_COLLECTION_ANALYTICS=analytics

# RabbitMQ Configuration
RABBITMQ_URL=amqp://admin:password@rabbitmq:5672/
RABBITMQ_QUEUE=video_analysis
RABBITMQ_EXCHANGE=video_events
RABBITMQ_ROUTING_KEY=analysis.request
```

#### Realtime Service

```bash
# Socket.IO Configuration
PORT=3001
HOST=0.0.0.0
CORS_ORIGIN=https://app.videoconverter.com,http://localhost:3000
SOCKET_IO_PATH=/socket.io/

# Connection Management
MAX_CONNECTIONS=10000
CONNECTION_TIMEOUT=60s
HEARTBEAT_INTERVAL=25s
HEARTBEAT_TIMEOUT=60s

# Authentication
JWT_SECRET_FILE=/etc/secrets/jwt-secret
AUTH_TIMEOUT=30s

# Redis Configuration
REDIS_URL=redis://redis:6379
REDIS_KEY_PREFIX=realtime:
REDIS_SUBSCRIPTION_CHANNELS=conversion:progress,conversion:complete,conversion:error

# Room Management
MAX_ROOMS=1000
ROOM_CLEANUP_INTERVAL=5m
USER_ROOM_PREFIX=user:

# Performance
CLUSTER_MODE=false                # Enable for multi-instance deployment
STICKY_SESSIONS=true
```

#### Frontend Service

```bash
# Next.js Configuration
NODE_ENV=production
PORT=3000
HOSTNAME=0.0.0.0

# API Configuration
NEXT_PUBLIC_API_URL=https://api.videoconverter.com/api/v1
NEXT_PUBLIC_REALTIME_URL=https://realtime.videoconverter.com
NEXT_PUBLIC_WS_RECONNECT_ATTEMPTS=5
NEXT_PUBLIC_WS_RECONNECT_DELAY=1000

# Authentication
NEXT_PUBLIC_JWT_STORAGE_KEY=video_converter_token
NEXT_PUBLIC_REFRESH_TOKEN_KEY=video_converter_refresh
NEXT_PUBLIC_TOKEN_REFRESH_THRESHOLD=300  # 5 minutes before expiry

# File Upload
NEXT_PUBLIC_MAX_FILE_SIZE=104857600      # 100MB
NEXT_PUBLIC_ALLOWED_FILE_TYPES=video/mp4,video/avi,video/mov,video/mkv
NEXT_PUBLIC_CHUNK_SIZE=1048576           # 1MB chunks

# UI Configuration
NEXT_PUBLIC_THEME=light
NEXT_PUBLIC_LANGUAGE=en
NEXT_PUBLIC_TIMEZONE=UTC
```

#### Notification Service

```bash
# SMTP Configuration
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=notifications@videoconverter.com
SMTP_PASSWORD_FILE=/etc/secrets/smtp-password
SMTP_TLS=true
SMTP_AUTH=true

# Email Templates
TEMPLATE_DIR=/app/templates
DEFAULT_FROM_EMAIL=noreply@videoconverter.com
DEFAULT_FROM_NAME=Video Converter

# Queue Configuration
RABBITMQ_URL=amqp://admin:password@rabbitmq:5672/
RABBITMQ_QUEUE=notifications
RABBITMQ_EXCHANGE=notifications
RABBITMQ_ROUTING_KEY=email.send

# Rate Limiting
EMAIL_RATE_LIMIT=100              # emails per hour
EMAIL_BURST_LIMIT=10              # burst emails

# Retry Configuration
MAX_RETRY_ATTEMPTS=3
RETRY_DELAY=5m
DEAD_LETTER_QUEUE=notifications.failed
```

## ConfigMaps and Secrets

### Creating ConfigMaps

#### Gateway Service ConfigMap

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: gateway-config
  namespace: video-converter
data:
  # Server Configuration
  PORT: '8080'
  CORS_ORIGINS: 'https://app.videoconverter.com'

  # File Upload
  UPLOAD_MAX_SIZE: '100MB'
  UPLOAD_ALLOWED_TYPES: 'video/mp4,video/avi,video/mov,video/mkv'

  # Rate Limiting
  RATE_LIMIT_ENABLED: 'true'
  RATE_LIMIT_REQUESTS_PER_MINUTE: '60'

  # Service URLs
  AUTH_SERVICE_URL: 'auth-service:50051'
  ANALYTICS_SERVICE_URL: 'analytics-service:50052'

  # Database
  MONGODB_URI: 'mongodb://mongodb:27017/video_converter'
  REDIS_URL: 'redis://redis:6379'
```

#### Analytics Service ConfigMap

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: analytics-config
  namespace: video-converter
data:
  # Server Configuration
  HOST: '0.0.0.0'
  PORT: '8000'
  WORKERS: '4'

  # ML Configuration
  MODEL_CACHE_SIZE: '3'
  GPU_ENABLED: 'false'

  # Video Processing
  MAX_VIDEO_DURATION: '3600'
  THUMBNAIL_COUNT: '5'
  THUMBNAIL_SIZE: '320x240'

  # Content Analysis
  CONTENT_SAFETY_THRESHOLD: '0.8'
  QUALITY_ANALYSIS_ENABLED: 'true'

  # Database
  MONGODB_URI: 'mongodb://mongodb:27017/video_converter'
```

### Creating Secrets

#### Database Secrets

```bash
kubectl create secret generic database-secrets \
  --from-literal=postgres-password=$(openssl rand -base64 32) \
  --from-literal=mongodb-password=$(openssl rand -base64 32) \
  --from-literal=redis-password=$(openssl rand -base64 32) \
  --namespace=video-converter
```

#### Application Secrets

```bash
kubectl create secret generic app-secrets \
  --from-literal=jwt-secret=$(openssl rand -base64 64) \
  --from-literal=smtp-password="your-smtp-password" \
  --namespace=video-converter
```

#### TLS Secrets

```bash
kubectl create secret tls video-converter-tls \
  --cert=path/to/tls.crt \
  --key=path/to/tls.key \
  --namespace=video-converter
```

### Using ConfigMaps and Secrets in Deployments

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: gateway-service
spec:
  template:
    spec:
      containers:
        - name: gateway
          image: video-converter/gateway:latest
          envFrom:
            - configMapRef:
                name: gateway-config
          env:
            - name: POSTGRES_PASSWORD
              valueFrom:
                secretKeyRef:
                  name: database-secrets
                  key: postgres-password
            - name: JWT_SECRET
              valueFrom:
                secretKeyRef:
                  name: app-secrets
                  key: jwt-secret
          volumeMounts:
            - name: tls-certs
              mountPath: /etc/tls
              readOnly: true
      volumes:
        - name: tls-certs
          secret:
            secretName: video-converter-tls
```

## Service Configuration

### gRPC Configuration

#### Server Configuration (Go)

```go
// internal/config/grpc.go
type GRPCConfig struct {
    Port                int           `env:"GRPC_PORT" envDefault:"50051"`
    Host                string        `env:"GRPC_HOST" envDefault:"0.0.0.0"`
    MaxConnections      int           `env:"GRPC_MAX_CONNECTIONS" envDefault:"1000"`
    ConnectionTimeout   time.Duration `env:"GRPC_CONNECTION_TIMEOUT" envDefault:"30s"`
    KeepAliveTime       time.Duration `env:"GRPC_KEEPALIVE_TIME" envDefault:"30s"`
    KeepAliveTimeout    time.Duration `env:"GRPC_KEEPALIVE_TIMEOUT" envDefault:"5s"`
    MaxRecvMsgSize      int           `env:"GRPC_MAX_RECV_MSG_SIZE" envDefault:"4194304"` // 4MB
    MaxSendMsgSize      int           `env:"GRPC_MAX_SEND_MSG_SIZE" envDefault:"4194304"` // 4MB
}

func (c *GRPCConfig) ServerOptions() []grpc.ServerOption {
    return []grpc.ServerOption{
        grpc.MaxRecvMsgSize(c.MaxRecvMsgSize),
        grpc.MaxSendMsgSize(c.MaxSendMsgSize),
        grpc.KeepaliveParams(keepalive.ServerParameters{
            Time:    c.KeepAliveTime,
            Timeout: c.KeepAliveTimeout,
        }),
        grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
            MinTime:             10 * time.Second,
            PermitWithoutStream: true,
        }),
    }
}
```

#### Client Configuration (Go)

```go
// internal/config/grpc_client.go
type GRPCClientConfig struct {
    AuthServiceURL      string        `env:"AUTH_SERVICE_URL" envDefault:"auth-service:50051"`
    AnalyticsServiceURL string        `env:"ANALYTICS_SERVICE_URL" envDefault:"analytics-service:50052"`
    Timeout             time.Duration `env:"GRPC_TIMEOUT" envDefault:"30s"`
    MaxRetryAttempts    int           `env:"GRPC_MAX_RETRY_ATTEMPTS" envDefault:"3"`
    RetryDelay          time.Duration `env:"GRPC_RETRY_DELAY" envDefault:"1s"`
}

func (c *GRPCClientConfig) DialOptions() []grpc.DialOption {
    return []grpc.DialOption{
        grpc.WithTransportCredentials(insecure.NewCredentials()),
        grpc.WithKeepaliveParams(keepalive.ClientParameters{
            Time:                10 * time.Second,
            Timeout:             time.Second,
            PermitWithoutStream: true,
        }),
        grpc.WithDefaultServiceConfig(`{
            "methodConfig": [{
                "name": [{"service": ""}],
                "retryPolicy": {
                    "MaxAttempts": 3,
                    "InitialBackoff": "1s",
                    "MaxBackoff": "10s",
                    "BackoffMultiplier": 2.0,
                    "RetryableStatusCodes": ["UNAVAILABLE", "DEADLINE_EXCEEDED"]
                }
            }]
        }`),
    }
}
```

### HTTP Configuration

#### Gin Server Configuration (Go)

```go
// internal/config/http.go
type HTTPConfig struct {
    Port            int           `env:"PORT" envDefault:"8080"`
    Host            string        `env:"HOST" envDefault:"0.0.0.0"`
    ReadTimeout     time.Duration `env:"HTTP_READ_TIMEOUT" envDefault:"30s"`
    WriteTimeout    time.Duration `env:"HTTP_WRITE_TIMEOUT" envDefault:"30s"`
    IdleTimeout     time.Duration `env:"HTTP_IDLE_TIMEOUT" envDefault:"120s"`
    MaxHeaderBytes  int           `env:"HTTP_MAX_HEADER_BYTES" envDefault:"1048576"` // 1MB
    CORSOrigins     []string      `env:"CORS_ORIGINS" envSeparator:","`
    CORSMethods     []string      `env:"CORS_METHODS" envSeparator:"," envDefault:"GET,POST,PUT,DELETE,OPTIONS"`
    CORSHeaders     []string      `env:"CORS_HEADERS" envSeparator:"," envDefault:"Content-Type,Authorization"`
}

func (c *HTTPConfig) Server(handler http.Handler) *http.Server {
    return &http.Server{
        Addr:           fmt.Sprintf("%s:%d", c.Host, c.Port),
        Handler:        handler,
        ReadTimeout:    c.ReadTimeout,
        WriteTimeout:   c.WriteTimeout,
        IdleTimeout:    c.IdleTimeout,
        MaxHeaderBytes: c.MaxHeaderBytes,
    }
}
```

### WebSocket Configuration

#### Socket.IO Configuration (TypeScript)

```typescript
// src/config/websocket.ts
export interface WebSocketConfig {
  port: number;
  host: string;
  corsOrigin: string[];
  path: string;
  maxConnections: number;
  connectionTimeout: number;
  heartbeatInterval: number;
  heartbeatTimeout: number;
  transports: string[];
}

export const getWebSocketConfig = (): WebSocketConfig => ({
  port: parseInt(process.env.PORT || '3001'),
  host: process.env.HOST || '0.0.0.0',
  corsOrigin: (process.env.CORS_ORIGIN || '').split(',').filter(Boolean),
  path: process.env.SOCKET_IO_PATH || '/socket.io/',
  maxConnections: parseInt(process.env.MAX_CONNECTIONS || '10000'),
  connectionTimeout: parseInt(process.env.CONNECTION_TIMEOUT || '60000'),
  heartbeatInterval: parseInt(process.env.HEARTBEAT_INTERVAL || '25000'),
  heartbeatTimeout: parseInt(process.env.HEARTBEAT_TIMEOUT || '60000'),
  transports: ['websocket', 'polling'],
});

export const createSocketIOServer = (config: WebSocketConfig) => {
  return new Server({
    cors: {
      origin: config.corsOrigin,
      methods: ['GET', 'POST'],
    },
    path: config.path,
    transports: config.transports,
    pingInterval: config.heartbeatInterval,
    pingTimeout: config.heartbeatTimeout,
    maxHttpBufferSize: 1e6, // 1MB
    allowEIO3: true,
  });
};
```

## Database Configuration

### PostgreSQL Configuration

#### Connection Configuration

```go
// internal/config/postgres.go
type PostgreSQLConfig struct {
    Host               string        `env:"POSTGRES_HOST" envDefault:"localhost"`
    Port               int           `env:"POSTGRES_PORT" envDefault:"5432"`
    Database           string        `env:"POSTGRES_DB" envDefault:"video_converter"`
    User               string        `env:"POSTGRES_USER" envDefault:"postgres"`
    Password           string        `env:"POSTGRES_PASSWORD"`
    SSLMode            string        `env:"POSTGRES_SSL_MODE" envDefault:"disable"`
    MaxConnections     int           `env:"POSTGRES_MAX_CONNECTIONS" envDefault:"50"`
    MaxIdleConnections int           `env:"POSTGRES_MAX_IDLE_CONNECTIONS" envDefault:"10"`
    ConnectionTimeout  time.Duration `env:"POSTGRES_CONNECTION_TIMEOUT" envDefault:"30s"`
    QueryTimeout       time.Duration `env:"POSTGRES_QUERY_TIMEOUT" envDefault:"30s"`
}

func (c *PostgreSQLConfig) DSN() string {
    return fmt.Sprintf(
        "host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
        c.Host, c.Port, c.User, c.Password, c.Database, c.SSLMode,
    )
}

func (c *PostgreSQLConfig) GormConfig() *gorm.Config {
    return &gorm.Config{
        Logger: logger.Default.LogMode(logger.Info),
        NamingStrategy: schema.NamingStrategy{
            SingularTable: false,
        },
    }
}
```

#### Production PostgreSQL Settings

```sql
-- postgresql.conf optimizations
shared_buffers = 256MB
effective_cache_size = 1GB
maintenance_work_mem = 64MB
checkpoint_completion_target = 0.9
wal_buffers = 16MB
default_statistics_target = 100
random_page_cost = 1.1
effective_io_concurrency = 200
work_mem = 4MB
min_wal_size = 1GB
max_wal_size = 4GB
max_worker_processes = 8
max_parallel_workers_per_gather = 4
max_parallel_workers = 8
max_parallel_maintenance_workers = 4
```

### MongoDB Configuration

#### Connection Configuration

```go
// internal/config/mongodb.go
type MongoDBConfig struct {
    URI            string        `env:"MONGODB_URI" envDefault:"mongodb://localhost:27017/video_converter"`
    Database       string        `env:"MONGODB_DATABASE" envDefault:"video_converter"`
    MaxPoolSize    uint64        `env:"MONGODB_MAX_POOL_SIZE" envDefault:"100"`
    MinPoolSize    uint64        `env:"MONGODB_MIN_POOL_SIZE" envDefault:"5"`
    Timeout        time.Duration `env:"MONGODB_TIMEOUT" envDefault:"30s"`
    GridFSBucket   string        `env:"GRIDFS_BUCKET" envDefault:"videos"`
    GridFSChunkSize int32        `env:"GRIDFS_CHUNK_SIZE" envDefault:"261120"` // 255KB
}

func (c *MongoDBConfig) ClientOptions() *options.ClientOptions {
    return options.Client().
        ApplyURI(c.URI).
        SetMaxPoolSize(c.MaxPoolSize).
        SetMinPoolSize(c.MinPoolSize).
        SetConnectTimeout(c.Timeout).
        SetServerSelectionTimeout(c.Timeout)
}
```

### Redis Configuration

#### Connection Configuration

```go
// internal/config/redis.go
type RedisConfig struct {
    URL            string        `env:"REDIS_URL" envDefault:"redis://localhost:6379"`
    MaxConnections int           `env:"REDIS_MAX_CONNECTIONS" envDefault:"100"`
    MinConnections int           `env:"REDIS_MIN_CONNECTIONS" envDefault:"10"`
    Timeout        time.Duration `env:"REDIS_TIMEOUT" envDefault:"5s"`
    KeyPrefix      string        `env:"REDIS_KEY_PREFIX" envDefault:"video_converter:"`
    DB             int           `env:"REDIS_DB" envDefault:"0"`
}

func (c *RedisConfig) Options() *redis.Options {
    opt, _ := redis.ParseURL(c.URL)
    opt.PoolSize = c.MaxConnections
    opt.MinIdleConns = c.MinConnections
    opt.DialTimeout = c.Timeout
    opt.ReadTimeout = c.Timeout
    opt.WriteTimeout = c.Timeout
    opt.DB = c.DB
    return opt
}
```

## Security Configuration

### TLS Configuration

#### Server TLS Configuration

```go
// internal/config/tls.go
type TLSConfig struct {
    Enabled  bool   `env:"TLS_ENABLED" envDefault:"false"`
    CertFile string `env:"TLS_CERT_FILE" envDefault:"/etc/tls/tls.crt"`
    KeyFile  string `env:"TLS_KEY_FILE" envDefault:"/etc/tls/tls.key"`
    CAFile   string `env:"TLS_CA_FILE"`
}

func (c *TLSConfig) ServerConfig() (*tls.Config, error) {
    if !c.Enabled {
        return nil, nil
    }

    cert, err := tls.LoadX509KeyPair(c.CertFile, c.KeyFile)
    if err != nil {
        return nil, err
    }

    config := &tls.Config{
        Certificates: []tls.Certificate{cert},
        MinVersion:   tls.VersionTLS12,
        CipherSuites: []uint16{
            tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
            tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
            tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
        },
    }

    if c.CAFile != "" {
        caCert, err := ioutil.ReadFile(c.CAFile)
        if err != nil {
            return nil, err
        }

        caCertPool := x509.NewCertPool()
        caCertPool.AppendCertsFromPEM(caCert)
        config.ClientCAs = caCertPool
        config.ClientAuth = tls.RequireAndVerifyClientCert
    }

    return config, nil
}
```

### JWT Configuration

#### JWT Settings

```go
// internal/config/jwt.go
type JWTConfig struct {
    Secret         string        `env:"JWT_SECRET"`
    Expiry         time.Duration `env:"JWT_EXPIRY" envDefault:"24h"`
    RefreshExpiry  time.Duration `env:"JWT_REFRESH_EXPIRY" envDefault:"168h"`
    Issuer         string        `env:"JWT_ISSUER" envDefault:"video-converter-auth"`
    Audience       string        `env:"JWT_AUDIENCE" envDefault:"video-converter-api"`
    Algorithm      string        `env:"JWT_ALGORITHM" envDefault:"HS256"`
}

func (c *JWTConfig) SigningMethod() jwt.SigningMethod {
    switch c.Algorithm {
    case "HS256":
        return jwt.SigningMethodHS256
    case "HS384":
        return jwt.SigningMethodHS384
    case "HS512":
        return jwt.SigningMethodHS512
    case "RS256":
        return jwt.SigningMethodRS256
    default:
        return jwt.SigningMethodHS256
    }
}
```

## Performance Tuning

### Resource Limits

#### Development Environment

```yaml
resources:
  requests:
    memory: '128Mi'
    cpu: '100m'
  limits:
    memory: '512Mi'
    cpu: '500m'
```

#### Staging Environment

```yaml
resources:
  requests:
    memory: '256Mi'
    cpu: '200m'
  limits:
    memory: '1Gi'
    cpu: '1000m'
```

#### Production Environment

```yaml
resources:
  requests:
    memory: '512Mi'
    cpu: '500m'
  limits:
    memory: '2Gi'
    cpu: '2000m'
```

### Horizontal Pod Autoscaler

```yaml
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
  behavior:
    scaleDown:
      stabilizationWindowSeconds: 300
      policies:
        - type: Percent
          value: 10
          periodSeconds: 60
    scaleUp:
      stabilizationWindowSeconds: 60
      policies:
        - type: Percent
          value: 50
          periodSeconds: 60
```

### JVM Tuning (for Java services)

```bash
# JVM Options
JAVA_OPTS="-Xms512m -Xmx2g -XX:+UseG1GC -XX:MaxGCPauseMillis=200 -XX:+UseStringDeduplication"
```

### Go Runtime Tuning

```bash
# Go Runtime Environment Variables
GOGC=100                    # GC target percentage
GOMAXPROCS=0               # Use all available CPUs
GOMEMLIMIT=2GiB            # Memory limit for GC
```

## Monitoring Configuration

### Prometheus Configuration

```yaml
# prometheus.yml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

rule_files:
  - 'alert_rules.yml'

scrape_configs:
  - job_name: 'kubernetes-pods'
    kubernetes_sd_configs:
      - role: pod
    relabel_configs:
      - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_scrape]
        action: keep
        regex: true
      - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_path]
        action: replace
        target_label: __metrics_path__
        regex: (.+)
      - source_labels: [__address__, __meta_kubernetes_pod_annotation_prometheus_io_port]
        action: replace
        regex: ([^:]+)(?::\d+)?;(\d+)
        replacement: $1:$2
        target_label: __address__

  - job_name: 'video-converter-services'
    static_configs:
      - targets:
          - 'gateway-service:8080'
          - 'auth-service:8080'
          - 'converter-service:8080'
          - 'analytics-service:8000'
          - 'realtime-service:3001'
          - 'notification-service:8080'
```

### Grafana Configuration

```yaml
# grafana.ini
[server]
http_port = 3000
domain = grafana.videoconverter.com

[security]
admin_user = admin
admin_password = ${GRAFANA_ADMIN_PASSWORD}

[auth]
disable_login_form = false

[auth.anonymous]
enabled = false

[dashboards]
default_home_dashboard_path = /var/lib/grafana/dashboards/overview.json

[alerting]
enabled = true
execute_alerts = true

[smtp]
enabled = true
host = ${SMTP_HOST}:${SMTP_PORT}
user = ${SMTP_USER}
password = ${SMTP_PASSWORD}
from_address = grafana@videoconverter.com
```

### Jaeger Configuration

```yaml
# jaeger-config.yml
apiVersion: v1
kind: ConfigMap
metadata:
  name: jaeger-config
data:
  jaeger.yml: |
    collector:
      zipkin:
        host-port: :9411
    storage:
      type: elasticsearch
      options:
        es:
          server-urls: http://elasticsearch:9200
          index-prefix: jaeger
    query:
      base-path: /jaeger
```

This configuration guide provides comprehensive coverage of all configuration aspects for the Video
Converter microservices platform. Use these configurations as templates and adjust them based on
your specific requirements and environment constraints.
