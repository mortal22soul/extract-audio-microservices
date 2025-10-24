#!/bin/bash

# Complete database setup script for video converter microservices
# This script sets up PostgreSQL, MongoDB, Redis, and RabbitMQ

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DOCKER_DIR="$SCRIPT_DIR/docker"
DATABASE_DIR="$SCRIPT_DIR/database"

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Video Converter Database Setup${NC}"
echo -e "${BLUE}========================================${NC}"

# Check if Docker and Docker Compose are installed
if ! command -v docker &> /dev/null; then
    echo -e "${RED}Docker is not installed. Please install Docker first.${NC}"
    exit 1
fi

if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
    echo -e "${RED}Docker Compose is not installed. Please install Docker Compose first.${NC}"
    exit 1
fi

# Determine Docker Compose command
if docker compose version &> /dev/null; then
    DOCKER_COMPOSE_CMD="docker compose"
else
    DOCKER_COMPOSE_CMD="docker-compose"
fi

echo -e "${GREEN}✓ Docker and Docker Compose are available${NC}"

# Function to wait for service to be healthy
wait_for_service() {
    local service_name=$1
    local max_attempts=30
    local attempt=1
    
    echo -e "${YELLOW}Waiting for $service_name to be healthy...${NC}"
    
    while [ $attempt -le $max_attempts ]; do
        if $DOCKER_COMPOSE_CMD -f "$DOCKER_DIR/databases.yml" ps "$service_name" | grep -q "healthy"; then
            echo -e "${GREEN}✓ $service_name is healthy${NC}"
            return 0
        fi
        
        echo -n "."
        sleep 2
        ((attempt++))
    done
    
    echo -e "${RED}✗ $service_name failed to become healthy within $((max_attempts * 2)) seconds${NC}"
    return 1
}

# Start database services
echo -e "${YELLOW}Starting database services...${NC}"
cd "$SCRIPT_DIR"
$DOCKER_COMPOSE_CMD -f docker/databases.yml up -d

# Wait for all services to be healthy
echo -e "${YELLOW}Waiting for services to start...${NC}"
sleep 10

wait_for_service "postgresql" || echo -e "${YELLOW}PostgreSQL may still be starting...${NC}"
wait_for_service "mongodb" || echo -e "${YELLOW}MongoDB may still be starting...${NC}"
wait_for_service "redis" || echo -e "${YELLOW}Redis may still be starting...${NC}"
wait_for_service "rabbitmq" || echo -e "${YELLOW}RabbitMQ may still be starting...${NC}"

# Additional wait to ensure all initialization scripts have run
echo -e "${YELLOW}Allowing time for initialization scripts...${NC}"
sleep 15

# Run database-specific setup scripts
echo -e "${YELLOW}Running database setup scripts...${NC}"

# PostgreSQL migrations
if [ -f "$DATABASE_DIR/postgresql/migrate.sh" ]; then
    echo -e "${YELLOW}Running PostgreSQL migrations...${NC}"
    cd "$DATABASE_DIR/postgresql"
    chmod +x migrate.sh
    ./migrate.sh up || echo -e "${YELLOW}PostgreSQL migrations may have already been applied${NC}"
fi

# Redis setup
if [ -f "$DATABASE_DIR/redis/setup.sh" ]; then
    echo -e "${YELLOW}Setting up Redis...${NC}"
    cd "$DATABASE_DIR/redis"
    chmod +x setup.sh
    ./setup.sh || echo -e "${YELLOW}Redis setup may have already been completed${NC}"
fi

# RabbitMQ setup
if [ -f "$DATABASE_DIR/rabbitmq/setup.sh" ]; then
    echo -e "${YELLOW}Setting up RabbitMQ...${NC}"
    cd "$DATABASE_DIR/rabbitmq"
    chmod +x setup.sh
    ./setup.sh || echo -e "${YELLOW}RabbitMQ setup may have already been completed${NC}"
fi

# Verify all services are running
echo -e "${YELLOW}Verifying service status...${NC}"
cd "$SCRIPT_DIR"
$DOCKER_COMPOSE_CMD -f docker/databases.yml ps

echo -e "${BLUE}========================================${NC}"
echo -e "${GREEN}Database Setup Complete!${NC}"
echo -e "${BLUE}========================================${NC}"

echo -e "${GREEN}Services are running on the following ports:${NC}"
echo -e "PostgreSQL:     ${YELLOW}localhost:5432${NC}"
echo -e "MongoDB:        ${YELLOW}localhost:27017${NC}"
echo -e "Redis:          ${YELLOW}localhost:6379${NC}"
echo -e "RabbitMQ AMQP:  ${YELLOW}localhost:5672${NC}"
echo -e "RabbitMQ Mgmt:  ${YELLOW}localhost:15672${NC}"

echo -e "\n${GREEN}Admin interfaces (use --profile admin-tools):${NC}"
echo -e "pgAdmin:        ${YELLOW}http://localhost:8080${NC} (admin@example.com / admin123)"
echo -e "Mongo Express:  ${YELLOW}http://localhost:8081${NC} (admin / admin123)"
echo -e "Redis Commander:${YELLOW}http://localhost:8082${NC} (admin / admin123)"
echo -e "RabbitMQ Mgmt:  ${YELLOW}http://localhost:15672${NC} (admin / admin_password_123)"

echo -e "\n${GREEN}Database credentials:${NC}"
echo -e "PostgreSQL: ${YELLOW}app_user / dev_password_123${NC}"
echo -e "MongoDB:    ${YELLOW}app_user / dev_password_123${NC}"
echo -e "Redis:      ${YELLOW}password: dev_redis_password_123${NC}"
echo -e "RabbitMQ:   ${YELLOW}app_user / app_password_123${NC}"

echo -e "\n${YELLOW}To start admin tools:${NC}"
echo -e "docker-compose -f docker/databases.yml --profile admin-tools up -d"

echo -e "\n${YELLOW}To stop all services:${NC}"
echo -e "docker-compose -f docker/databases.yml down"

echo -e "\n${YELLOW}To view logs:${NC}"
echo -e "docker-compose -f docker/databases.yml logs -f [service_name]"