#!/bin/bash

# Redis setup script for video converter microservices
# This script initializes Redis with the required data structures and configurations

set -e

# Configuration
REDIS_HOST=${REDIS_HOST:-localhost}
REDIS_PORT=${REDIS_PORT:-6379}
REDIS_PASSWORD=${REDIS_PASSWORD:-dev_redis_password_123}

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}Setting up Redis for video converter microservices...${NC}"

# Check if redis-cli is available
if ! command -v redis-cli &> /dev/null; then
    echo -e "${RED}redis-cli is not installed. Please install Redis client tools.${NC}"
    exit 1
fi

# Test Redis connection
echo -e "${YELLOW}Testing Redis connection...${NC}"
if ! redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" -a "$REDIS_PASSWORD" ping > /dev/null 2>&1; then
    echo -e "${RED}Cannot connect to Redis at $REDIS_HOST:$REDIS_PORT${NC}"
    echo "Please ensure Redis is running and the credentials are correct."
    exit 1
fi

echo -e "${GREEN}Redis connection successful!${NC}"

# Load Lua initialization script
echo -e "${YELLOW}Loading Redis initialization script...${NC}"
redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" -a "$REDIS_PASSWORD" --eval init.lua

# Set up pub/sub channels
echo -e "${YELLOW}Setting up pub/sub channels...${NC}"
redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" -a "$REDIS_PASSWORD" << 'EOF'
# Create channel subscriptions info
HSET channels:info conversion:progress "Video conversion progress updates"
HSET channels:info conversion:complete "Video conversion completion notifications"
HSET channels:info conversion:error "Video conversion error notifications"
HSET channels:info user:activity "User activity tracking"
HSET channels:info system:alerts "System-wide alerts and notifications"
EOF

# Set up Redis data structures for different use cases
echo -e "${YELLOW}Setting up Redis data structures...${NC}"
redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" -a "$REDIS_PASSWORD" << 'EOF'
# Configuration settings
HSET config:app max_file_size 104857600
HSET config:app supported_formats "mp4,avi,mov,mkv,webm,flv"
HSET config:app max_conversion_time 3600
HSET config:app max_concurrent_jobs 5
HSET config:app cleanup_interval 86400

# Rate limiting configurations
HSET config:rate_limits upload_per_hour 10
HSET config:rate_limits api_requests_per_minute 100
HSET config:rate_limits conversion_per_day 50

# Cache TTL settings
HSET config:ttl user_session 86400
HSET config:ttl conversion_progress 3600
HSET config:ttl user_activity 604800
HSET config:ttl rate_limit 3600

# System status
HSET system:status redis_initialized 1
HSET system:status last_setup_time $(date +%s)
HSET system:status version "1.0.0"
EOF

# Create indexes for sorted sets (for leaderboards, rankings, etc.)
echo -e "${YELLOW}Setting up sorted sets for analytics...${NC}"
redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" -a "$REDIS_PASSWORD" << 'EOF'
# Video conversion statistics
ZADD stats:conversions:daily $(date +%s) 0
ZADD stats:uploads:daily $(date +%s) 0
ZADD stats:users:active $(date +%s) 0

# Popular video formats
ZADD stats:formats 0 mp4
ZADD stats:formats 0 avi
ZADD stats:formats 0 mov
ZADD stats:formats 0 mkv
ZADD stats:formats 0 webm
EOF

# Set up connection pool configuration
echo -e "${YELLOW}Configuring connection pool settings...${NC}"
redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" -a "$REDIS_PASSWORD" << 'EOF'
# Connection pool settings for different services
HSET pool:auth max_connections 20
HSET pool:auth min_connections 5
HSET pool:auth connection_timeout 5000

HSET pool:gateway max_connections 50
HSET pool:gateway min_connections 10
HSET pool:gateway connection_timeout 5000

HSET pool:converter max_connections 30
HSET pool:converter min_connections 5
HSET pool:converter connection_timeout 10000

HSET pool:realtime max_connections 100
HSET pool:realtime min_connections 20
HSET pool:realtime connection_timeout 3000
EOF

# Verify setup
echo -e "${YELLOW}Verifying Redis setup...${NC}"
redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" -a "$REDIS_PASSWORD" << 'EOF'
# Check if all configurations are set
EXISTS config:app
EXISTS config:rate_limits
EXISTS config:ttl
EXISTS system:status
HGETALL system:status
EOF

echo -e "${GREEN}Redis setup completed successfully!${NC}"
echo -e "${GREEN}Pub/sub channels configured for real-time communication${NC}"
echo -e "${GREEN}Data structures initialized for caching and session management${NC}"
echo -e "${GREEN}Rate limiting and connection pooling configured${NC}"

# Display connection info
echo -e "\n${YELLOW}Redis Connection Information:${NC}"
echo "Host: $REDIS_HOST"
echo "Port: $REDIS_PORT"
echo "Password: [CONFIGURED]"
echo -e "\n${YELLOW}Available Pub/Sub Channels:${NC}"
echo "- conversion:progress"
echo "- conversion:complete"
echo "- conversion:error"
echo "- user:activity"
echo "- system:alerts"