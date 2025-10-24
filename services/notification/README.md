# Notification Service

A Go-based email notification service that consumes messages from RabbitMQ and sends emails via SMTP
with support for multiple providers.

## Features

- **Multi-provider SMTP support**: Gmail, Outlook, Yahoo, SendGrid, Mailgun, and custom SMTP
- **Template-based emails**: HTML and text templates with built-in templates for common
  notifications
- **RabbitMQ integration**: Consumes notification jobs from message queue with retry logic
- **User preferences**: Configurable notification preferences and unsubscribe functionality
- **Health checks**: HTTP endpoints for health, readiness, and liveness probes
- **Retry logic**: Automatic retry for failed email deliveries and RabbitMQ connections
- **Dead letter queues**: Failed messages are sent to dead letter queue for analysis

## Architecture

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   RabbitMQ      │───▶│ Notification     │───▶│   SMTP          │
│   Queue         │    │ Service          │    │   Provider      │
└─────────────────┘    └──────────────────┘    └─────────────────┘
                              │
                              ▼
                       ┌──────────────────┐
                       │ Template Engine  │
                       │ (HTML/Text)      │
                       └──────────────────┘
```

## Setup

### 1. Environment Configuration

Copy the example environment file and configure your settings:

```bash
cp .env.example .env
```

Edit `.env` with your SMTP provider settings:

```env
# RabbitMQ Configuration
RABBITMQ_URL=amqp://guest:guest@localhost:5672/
NOTIFICATION_QUEUE=notifications

# SMTP Configuration (Gmail example)
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your-email@gmail.com
SMTP_PASSWORD=your-app-password
SMTP_FROM=noreply@videoconverter.com

# Service Configuration
PORT=8080
```

### 2. SMTP Provider Setup

#### Gmail

1. Enable 2-factor authentication
2. Generate an app password
3. Use `smtp.gmail.com:587`

#### Outlook

1. Use `smtp-mail.outlook.com:587`
2. Use your regular password or app password

#### SendGrid

1. Create an API key
2. Use `smtp.sendgrid.net:587`
3. Username: `apikey`, Password: your API key

### 3. Dependencies

Install Go dependencies:

```bash
go mod tidy
```

### 4. Running the Service

```bash
# Development
go run cmd/main.go

# Build and run
go build -o notification cmd/main.go
./notification
```

## API Endpoints

### Health Checks

- `GET /health` - General health status
- `GET /health/ready` - Readiness probe
- `GET /health/live` - Liveness probe
- `GET /metrics` - Basic metrics

### Notification Management

- `GET /unsubscribe?token=<token>` - Unsubscribe from emails
- `GET /preferences/{userID}` - Get user notification preferences
- `PUT /preferences/{userID}` - Update user notification preferences

### Test Endpoints (Development)

- `POST /test/welcome` - Send test welcome email
- `POST /test/conversion-complete` - Send test conversion complete email
- `POST /test/conversion-error` - Send test conversion error email

## Message Format

The service expects RabbitMQ messages in the following JSON format:

```json
{
  "id": "notif_abc123_1234567890",
  "type": "welcome|conversion_complete|conversion_error",
  "user_id": "user123",
  "email": "user@example.com",
  "data": {
    "FirstName": "John",
    "VideoName": "my-video.mp4",
    "Duration": "5:30",
    "FileSize": "25.6 MB",
    "ConversionTime": "2m 15s",
    "ErrorMessage": "Unsupported format"
  },
  "priority": 1
}
```

## Email Templates

### Built-in Templates

1. **Welcome Email** - Sent to new users
2. **Conversion Complete** - Sent when video conversion succeeds
3. **Conversion Error** - Sent when video conversion fails

### Template Data

Each template receives specific data:

#### Welcome Email

- `FirstName` - User's first name
- `DashboardURL` - Link to user dashboard
- `UnsubscribeURL` - Unsubscribe link

#### Conversion Complete

- `FirstName` - User's first name
- `VideoName` - Original video filename
- `Duration` - Video duration
- `FileSize` - File size
- `ConversionTime` - Time taken for conversion
- `Quality` - Output quality (e.g., "192 kbps MP3")
- `DownloadURL` - Download link for converted file
- `DashboardURL` - Link to user dashboard
- `UnsubscribeURL` - Unsubscribe link

#### Conversion Error

- `FirstName` - User's first name
- `VideoName` - Original video filename
- `ErrorMessage` - Error description
- `ErrorTime` - When the error occurred
- `JobID` - Conversion job ID for support
- `UploadURL` - Link to upload page
- `SupportURL` - Link to support
- `UnsubscribeURL` - Unsubscribe link

## Testing

### Unit Tests

```bash
go test ./...
```

### Integration Testing

1. Start RabbitMQ and configure SMTP
2. Send test messages via HTTP endpoints:

```bash
# Test welcome email
curl -X POST http://localhost:8080/test/welcome \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "test123",
    "email": "test@example.com",
    "first_name": "Test User"
  }'

# Test conversion complete email
curl -X POST http://localhost:8080/test/conversion-complete \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "test123",
    "email": "test@example.com",
    "first_name": "Test User",
    "video_name": "test-video.mp4",
    "job_id": "job123",
    "duration": "3:45",
    "file_size": "15.2 MB",
    "conversion_time": "1m 30s"
  }'
```

## Docker

### Development Dockerfile

```dockerfile
FROM golang:1.21-alpine AS dev

WORKDIR /app
RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o notification cmd/main.go

EXPOSE 8080
CMD ["./notification"]
```

### Build and Run

```bash
docker build -f Dockerfile.dev -t notification-service .
docker run -p 8080:8080 --env-file .env notification-service
```

## Monitoring

### Health Checks

The service provides Kubernetes-compatible health check endpoints:

```yaml
livenessProbe:
  httpGet:
    path: /health/live
    port: 8080
  initialDelaySeconds: 30
  periodSeconds: 10

readinessProbe:
  httpGet:
    path: /health/ready
    port: 8080
  initialDelaySeconds: 5
  periodSeconds: 5
```

### Metrics

Basic metrics are available at `/metrics` endpoint:

- Service uptime
- SMTP provider information
- Notification counts (sent/failed)

## Troubleshooting

### Common Issues

1. **SMTP Authentication Failed**
   - Check username/password
   - For Gmail, use app passwords
   - Verify 2FA is enabled for Gmail

2. **RabbitMQ Connection Failed**
   - Check RabbitMQ URL
   - Verify RabbitMQ is running
   - Check network connectivity

3. **Templates Not Rendering**
   - Check template data format
   - Verify all required fields are provided
   - Check logs for template errors

### Logs

The service provides structured logging:

```
2024/01/01 12:00:00 Starting Notification Service...
2024/01/01 12:00:01 SMTP connection test successful - Provider: Gmail
2024/01/01 12:00:02 Connected to RabbitMQ on attempt 1
2024/01/01 12:00:03 Notification consumer started successfully
2024/01/01 12:00:04 Starting HTTP server on port 8080
```

## Production Considerations

1. **Security**
   - Use secure SMTP passwords/API keys
   - Implement proper authentication for HTTP endpoints
   - Use TLS for all connections

2. **Scalability**
   - Run multiple instances for high availability
   - Use persistent storage for user preferences
   - Implement proper metrics and monitoring

3. **Reliability**
   - Configure dead letter queues
   - Implement circuit breakers for SMTP
   - Add comprehensive logging and alerting
