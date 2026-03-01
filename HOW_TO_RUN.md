# HOW TO RUN

## Prerequisites

| Tool                    | Version      | Purpose                                |
| ----------------------- | ------------ | -------------------------------------- |
| Docker + Docker Compose | 24+          | Run all services locally               |
| Go                      | 1.21+        | Build gateway / converter / auth       |
| Node.js + pnpm          | 18+ / 8+     | Frontend + realtime service            |
| Python + uv             | 3.11+ / 0.4+ | Analytics service                      |
| kubectl                 | any          | Kubernetes deployments                 |
| Tilt                    | 0.33+        | Local Kubernetes dev loop _(optional)_ |

---

## Option 1 — Docker Compose (Recommended for local dev)

The fastest way to get everything running.

```bash
# 1. Copy and fill in environment variables
cp .env.example .env

# 2. Start all infrastructure + services
docker compose up -d

# 3. Watch logs (optional)
docker compose logs -f gateway-service converter-service
```

> First run will build images — takes ~3-5 minutes.

### Service Endpoints

| Service                  | URL                                           |
| ------------------------ | --------------------------------------------- |
| **Gateway API**          | http://localhost:8080                         |
| **Frontend**             | http://localhost:3000                         |
| **Analytics API**        | http://localhost:8000                         |
| **Realtime (WebSocket)** | http://localhost:3001                         |
| **MinIO Console**        | http://localhost:9001 (`admin` / `dev123456`) |
| **RabbitMQ Management**  | http://localhost:15672 (`admin` / `dev123`)   |

### Stop

```bash
docker compose down          # stop containers, keep volumes
docker compose down -v       # stop + wipe all data volumes
```

---

## Option 2 — Run Services Individually

Useful when you are actively developing a single service.

### 1. Start Infrastructure First

```bash
docker compose up -d postgresql mongodb redis rabbitmq minio
```

Wait for all health checks to pass:

```bash
docker compose ps
```

### 2. Gateway Service (Go)

```bash
cd services/gateway

export PORT=8080
export MONGODB_URI=mongodb://admin:dev123@localhost:27017/videoconverter?authSource=admin
export REDIS_URL=redis://:dev123@localhost:6379
export AUTH_SERVICE_URL=localhost:50051
export MINIO_ENDPOINT=localhost:9000
export MINIO_ACCESS_KEY=admin
export MINIO_SECRET_KEY=dev123456
export MINIO_BUCKET=videos
export MINIO_USE_SSL=false
export JWT_SECRET=dev-jwt-secret
export LOG_LEVEL=debug

go run ./cmd/main.go
```

### 3. Converter Service (Go)

```bash
cd services/converter

export MONGODB_URI=mongodb://admin:dev123@localhost:27017/videoconverter?authSource=admin
export RABBITMQ_URL=amqp://admin:dev123@localhost:5672/
export REDIS_URL=redis://:dev123@localhost:6379
export MINIO_ENDPOINT=localhost:9000
export MINIO_ACCESS_KEY=admin
export MINIO_SECRET_KEY=dev123456
export MINIO_BUCKET=videos
export MINIO_USE_SSL=false
export LOG_LEVEL=debug

go run ./cmd/main.go
```

> FFmpeg must be installed: `winget install Gyan.FFmpeg` (Windows) or `brew install ffmpeg` (macOS)

### 4. Auth Service (Go)

```bash
cd services/auth

export DATABASE_URL=postgres://postgres:dev123@localhost:5432/videoconverter?sslmode=disable
export GRPC_PORT=50051
export JWT_SECRET=dev-jwt-secret
export LOG_LEVEL=debug

go run ./cmd/main.go
```

### 5. Analytics Service (Python)

```bash
cd services/analytics

uv sync                      # install dependencies
uv run uvicorn src.analytics.api:app --host 0.0.0.0 --port 8000 --reload
```

### 6. Frontend (Next.js)

```bash
cd services/frontend

pnpm install
pnpm dev
```

### 7. Realtime Service (Node.js)

```bash
cd services/realtime

pnpm install
pnpm dev
```

---

## Option 3 — Kubernetes with Tilt

For a production-like local environment using Kubernetes.

```bash
# Requires: Docker Desktop (with Kubernetes enabled) or Minikube
tilt up
```

Tilt will:

- Apply all manifests in `infrastructure/k8s/` (databases, storage, monitoring)
- Build and deploy Go services with live-reload on file changes
- Open the Tilt dashboard at http://localhost:10350

```bash
tilt down   # tear everything down
```

---

## Verify the Stack is Working

### Quick health check

```bash
# Gateway
curl http://localhost:8080/health

# Analytics
curl http://localhost:8000/health
```

### Upload a video (end-to-end test)

```bash
curl -X POST http://localhost:8080/api/v1/videos/upload \
  -H "Authorization: Bearer <token>" \
  -F "video=@/path/to/video.mp4"
```

1. Video appears in MinIO → `videos/{id}/original.mp4`
2. Converter picks it up from RabbitMQ queue
3. FFmpeg extracts audio
4. MP3 uploaded to MinIO → `videos/{id}/output.mp3`
5. Realtime service pushes completion event via WebSocket

---

## Troubleshooting

| Problem                        | Fix                                                    |
| ------------------------------ | ------------------------------------------------------ |
| `minio: connection refused`    | Wait for MinIO health check: `docker compose ps minio` |
| `rabbitmq: connection refused` | RabbitMQ takes ~15s to start, retry after              |
| Go build errors                | Run `go mod tidy` in the affected service directory    |
| Python import errors           | Run `uv sync` in `services/analytics/`                 |
| Port already in use            | `docker compose down` then retry                       |
