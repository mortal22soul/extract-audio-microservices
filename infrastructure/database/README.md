# Database Infrastructure

This directory contains the complete database infrastructure setup for the video converter
microservices system. The infrastructure includes PostgreSQL, MongoDB, Redis, and RabbitMQ with
proper configuration, initialization scripts, and Docker Compose orchestration.

## Architecture Overview

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   PostgreSQL    │    │     MongoDB     │    │      Redis      │    │    RabbitMQ     │
│                 │    │                 │    │                 │    │                 │
│ User Auth       │    │ Video Metadata  │    │ Cache & PubSub  │    │ Message Queue   │
│ Sessions        │    │ Conversion Jobs │    │ Sessions        │    │ Async Tasks     │
│ Structured Data │    │ Analytics Data  │    │ Rate Limiting   │    │ Dead Letters    │
└─────────────────┘    └─────────────────┘    └─────────────────┘    └─────────────────┘
```

## Services

### PostgreSQL (Port 5432)

- **Purpose**: User authentication, sessions, and structured data
- **Database**: `video_converter_auth`
- **User**: `app_user` / `dev_password_123`
- **Features**:
  - User authentication schema with bcrypt password hashing
  - JWT session tracking
  - Database migrations with golang-migrate
  - Proper indexes for performance
  - Seed data for development

### MongoDB (Port 27017)

- **Purpose**: Video metadata and conversion job tracking
- **Database**: `video_converter`
- **User**: `app_user` / `dev_password_123`
- **Features**:
  - Video metadata and analytics data
  - Conversion job tracking
  - Schema validation with MongoDB validators
  - Optimized indexes for file queries

### Redis (Port 6379)

- **Purpose**: Caching, session storage, and pub/sub messaging
- **Password**: `dev_redis_password_123`
- **Features**:
  - User session caching
  - Real-time pub/sub for conversion progress
  - Rate limiting data structures
  - Connection pooling configuration
  - Persistence with RDB and AOF

### RabbitMQ (Ports 5672, 15672)

- **Purpose**: Message broker for asynchronous task processing
- **User**: `admin` / `admin_password_123`
- **App User**: `app_user` / `app_password_123`
- **Features**:
  - Topic and direct exchanges for different message types
  - Dead letter queues for error handling
  - Message TTL and retry policies
  - High availability configuration
  - Management UI for monitoring

## Quick Start

### Using Docker Compose (Recommended)

1. **Start all database services:**

   ```bash
   # Linux/macOS
   ./setup-databases.sh

   # Windows
   setup-databases.bat
   ```

2. **Or manually with Docker Compose:**

   ```bash
   docker-compose -f docker/databases.yml up -d
   ```

3. **Start with admin tools:**
   ```bash
   docker-compose -f docker/databases.yml --profile admin-tools up -d
   ```

### Individual Service Setup

Each database service can be started individually:

```bash
# PostgreSQL only
docker-compose -f docker/postgresql.yml up -d

# MongoDB only
docker-compose -f docker/mongodb.yml up -d

# Redis only
docker-compose -f docker/redis.yml up -d

# RabbitMQ only
docker-compose -f docker/rabbitmq.yml up -d
```

## Database Schemas

### PostgreSQL Schema

```sql
-- Users table
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- User sessions table
CREATE TABLE user_sessions (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(255) NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);
```

### MongoDB Collections

- **videos**: Video metadata and processing status
- **conversion_jobs**: Video processing job tracking
- **analytics_data**: ML analysis results

### Redis Data Structures

- **user:session:{userId}**: User session data (Hash)
- **conversion:{jobId}**: Conversion progress (Hash)
- **user:activity:{userId}**: User activity tracking (Sorted Set)
- **rate_limit:{userId}:{endpoint}**: Rate limiting (String with TTL)

### RabbitMQ Queues

- **video.processing.queue**: Video upload processing
- **video.conversion.queue**: Video-to-MP3 conversion
- **analytics.processing.queue**: ML analysis tasks
- **notification.email.queue**: Email notifications
- **dlx.failed.queue**: Dead letter queue for failed messages

## Management and Monitoring

### Admin Interfaces

When started with `--profile admin-tools`:

- **pgAdmin**: http://localhost:8080 (admin@example.com / admin123)
- **Mongo Express**: http://localhost:8081 (admin / admin123)
- **Redis Commander**: http://localhost:8082 (admin / admin123)
- **RabbitMQ Management**: http://localhost:15672 (admin / admin_password_123)

### Health Checks

All services include health checks:

```bash
# Check service status
docker-compose -f docker/databases.yml ps

# View service logs
docker-compose -f docker/databases.yml logs -f [service_name]
```

## Database Migrations

### PostgreSQL Migrations

```bash
cd database/postgresql

# Run all pending migrations
./migrate.sh up

# Rollback last migration
./migrate.sh down 1

# Check current version
./migrate.sh version

# Create new migration
./migrate.sh create add_new_feature
```

### MongoDB Migrations

MongoDB uses initialization scripts that run automatically:

- `init.js`: Creates collections, indexes, and users
- `seed.js`: Inserts sample data

## Configuration

### Environment Variables

```bash
# PostgreSQL
DB_HOST=localhost
DB_PORT=5432
DB_NAME=video_converter_auth
DB_USER=app_user
DB_PASSWORD=dev_password_123

# MongoDB
MONGO_HOST=localhost
MONGO_PORT=27017
MONGO_DB=video_converter
MONGO_USER=app_user
MONGO_PASSWORD=dev_password_123

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=dev_redis_password_123

# RabbitMQ
RABBITMQ_HOST=localhost
RABBITMQ_PORT=5672
RABBITMQ_MGMT_PORT=15672
RABBITMQ_USER=app_user
RABBITMQ_PASSWORD=app_password_123
RABBITMQ_VHOST=video_converter
```

### Production Considerations

For production deployment:

1. **Security**:
   - Change all default passwords
   - Enable SSL/TLS for all connections
   - Configure proper firewall rules
   - Use secrets management

2. **Performance**:
   - Tune memory settings for each service
   - Configure connection pooling
   - Set up monitoring and alerting
   - Implement backup strategies

3. **High Availability**:
   - Set up database replication
   - Configure RabbitMQ clustering
   - Use Redis Sentinel or Cluster
   - Implement load balancing

## Troubleshooting

### Common Issues

1. **Services not starting**:

   ```bash
   # Check Docker daemon
   docker info

   # Check service logs
   docker-compose -f docker/databases.yml logs [service_name]
   ```

2. **Connection refused**:
   - Ensure services are healthy: `docker-compose ps`
   - Check port conflicts: `netstat -tulpn | grep [port]`
   - Verify firewall settings

3. **Data persistence**:
   - Check Docker volumes: `docker volume ls`
   - Verify mount points in containers

4. **Performance issues**:
   - Monitor resource usage: `docker stats`
   - Check service-specific logs for errors
   - Review configuration settings

### Cleanup

```bash
# Stop all services
docker-compose -f docker/databases.yml down

# Remove volumes (WARNING: This deletes all data)
docker-compose -f docker/databases.yml down -v

# Remove images
docker-compose -f docker/databases.yml down --rmi all
```

## Development Workflow

1. **Initial Setup**: Run `setup-databases.sh` to start all services
2. **Development**: Services auto-restart and maintain data in volumes
3. **Testing**: Use seed data or create test fixtures
4. **Schema Changes**: Create migrations for PostgreSQL, update MongoDB init scripts
5. **Cleanup**: Use `docker-compose down` to stop services

## Support

For issues or questions:

1. Check service logs for error messages
2. Verify configuration files
3. Ensure all required ports are available
4. Review Docker and Docker Compose documentation
