#!/bin/bash

# RabbitMQ setup script for video converter microservices
# This script configures exchanges, queues, bindings, and policies

set -e

# Configuration
RABBITMQ_HOST=${RABBITMQ_HOST:-localhost}
RABBITMQ_PORT=${RABBITMQ_PORT:-15672}
RABBITMQ_USER=${RABBITMQ_USER:-admin}
RABBITMQ_PASSWORD=${RABBITMQ_PASSWORD:-admin_password_123}
RABBITMQ_VHOST=${RABBITMQ_VHOST:-video_converter}

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}Setting up RabbitMQ for video converter microservices...${NC}"

# Check if rabbitmqadmin is available
if ! command -v rabbitmqadmin &> /dev/null; then
    echo -e "${YELLOW}rabbitmqadmin not found. Downloading...${NC}"
    curl -o rabbitmqadmin http://$RABBITMQ_HOST:$RABBITMQ_PORT/cli/rabbitmqadmin
    chmod +x rabbitmqadmin
    RABBITMQADMIN_CMD="./rabbitmqadmin"
else
    RABBITMQADMIN_CMD="rabbitmqadmin"
fi

# Base command with authentication
RABBITMQ_CMD="$RABBITMQADMIN_CMD -H $RABBITMQ_HOST -P $RABBITMQ_PORT -u $RABBITMQ_USER -p $RABBITMQ_PASSWORD"

# Test RabbitMQ connection
echo -e "${YELLOW}Testing RabbitMQ connection...${NC}"
if ! $RABBITMQ_CMD list vhosts > /dev/null 2>&1; then
    echo -e "${RED}Cannot connect to RabbitMQ at $RABBITMQ_HOST:$RABBITMQ_PORT${NC}"
    echo "Please ensure RabbitMQ is running and the credentials are correct."
    exit 1
fi

echo -e "${GREEN}RabbitMQ connection successful!${NC}"

# Create virtual host if it doesn't exist
echo -e "${YELLOW}Creating virtual host: $RABBITMQ_VHOST${NC}"
$RABBITMQ_CMD declare vhost name=$RABBITMQ_VHOST || echo "Virtual host may already exist"

# Set permissions for the virtual host
echo -e "${YELLOW}Setting permissions for virtual host...${NC}"
rabbitmqctl set_permissions -p $RABBITMQ_VHOST $RABBITMQ_USER ".*" ".*" ".*" || echo "Permissions may already be set"

# Create exchanges
echo -e "${YELLOW}Creating exchanges...${NC}"
$RABBITMQ_CMD -V $RABBITMQ_VHOST declare exchange name=video.exchange type=topic durable=true
$RABBITMQ_CMD -V $RABBITMQ_VHOST declare exchange name=analytics.exchange type=topic durable=true
$RABBITMQ_CMD -V $RABBITMQ_VHOST declare exchange name=notification.exchange type=direct durable=true
$RABBITMQ_CMD -V $RABBITMQ_VHOST declare exchange name=dlx.exchange type=direct durable=true

# Create queues with dead letter exchange configuration
echo -e "${YELLOW}Creating queues...${NC}"

# Video processing queue
$RABBITMQ_CMD -V $RABBITMQ_VHOST declare queue name=video.processing.queue durable=true \
    arguments='{"x-message-ttl":3600000,"x-dead-letter-exchange":"dlx.exchange","x-dead-letter-routing-key":"video.failed","x-max-retries":3}'

# Video conversion queue
$RABBITMQ_CMD -V $RABBITMQ_VHOST declare queue name=video.conversion.queue durable=true \
    arguments='{"x-message-ttl":7200000,"x-dead-letter-exchange":"dlx.exchange","x-dead-letter-routing-key":"conversion.failed","x-max-retries":2}'

# Analytics processing queue
$RABBITMQ_CMD -V $RABBITMQ_VHOST declare queue name=analytics.processing.queue durable=true \
    arguments='{"x-message-ttl":1800000,"x-dead-letter-exchange":"dlx.exchange","x-dead-letter-routing-key":"analytics.failed","x-max-retries":3}'

# Notification email queue
$RABBITMQ_CMD -V $RABBITMQ_VHOST declare queue name=notification.email.queue durable=true \
    arguments='{"x-message-ttl":600000,"x-dead-letter-exchange":"dlx.exchange","x-dead-letter-routing-key":"notification.failed","x-max-retries":5}'

# Dead letter queue
$RABBITMQ_CMD -V $RABBITMQ_VHOST declare queue name=dlx.failed.queue durable=true

# Create bindings
echo -e "${YELLOW}Creating bindings...${NC}"

# Video exchange bindings
$RABBITMQ_CMD -V $RABBITMQ_VHOST declare binding source=video.exchange destination=video.processing.queue routing_key=video.uploaded
$RABBITMQ_CMD -V $RABBITMQ_VHOST declare binding source=video.exchange destination=video.conversion.queue routing_key=video.convert
$RABBITMQ_CMD -V $RABBITMQ_VHOST declare binding source=video.exchange destination=analytics.processing.queue routing_key=video.analyze

# Analytics exchange bindings
$RABBITMQ_CMD -V $RABBITMQ_VHOST declare binding source=analytics.exchange destination=analytics.processing.queue routing_key="analytics.*"

# Notification exchange bindings
$RABBITMQ_CMD -V $RABBITMQ_VHOST declare binding source=notification.exchange destination=notification.email.queue routing_key=email

# Dead letter exchange bindings
$RABBITMQ_CMD -V $RABBITMQ_VHOST declare binding source=dlx.exchange destination=dlx.failed.queue routing_key=failed

# Set up policies
echo -e "${YELLOW}Setting up policies...${NC}"

# High availability policy
rabbitmqctl set_policy -p $RABBITMQ_VHOST ha-all ".*" '{"ha-mode":"all","ha-sync-mode":"automatic"}' --priority 0 || echo "HA policy may already exist"

# Dead letter exchange policy for specific queues
rabbitmqctl set_policy -p $RABBITMQ_VHOST dlx-policy "^(video|notification|analytics)\." \
    '{"dead-letter-exchange":"dlx.exchange","dead-letter-routing-key":"failed","message-ttl":3600000}' --priority 1 || echo "DLX policy may already exist"

# Create application user
echo -e "${YELLOW}Creating application user...${NC}"
rabbitmqctl add_user app_user app_password_123 || echo "User may already exist"
rabbitmqctl set_user_tags app_user management || echo "User tags may already be set"
rabbitmqctl set_permissions -p $RABBITMQ_VHOST app_user ".*" ".*" ".*" || echo "User permissions may already be set"

# Verify setup
echo -e "${YELLOW}Verifying RabbitMQ setup...${NC}"
echo "Exchanges:"
$RABBITMQ_CMD -V $RABBITMQ_VHOST list exchanges name type durable

echo -e "\nQueues:"
$RABBITMQ_CMD -V $RABBITMQ_VHOST list queues name messages durable

echo -e "\nBindings:"
$RABBITMQ_CMD -V $RABBITMQ_VHOST list bindings source_name destination_name routing_key

echo -e "\n${GREEN}RabbitMQ setup completed successfully!${NC}"
echo -e "${GREEN}Exchanges, queues, and bindings configured${NC}"
echo -e "${GREEN}Dead letter queues set up for error handling${NC}"
echo -e "${GREEN}High availability and retry policies applied${NC}"

# Display connection info
echo -e "\n${YELLOW}RabbitMQ Connection Information:${NC}"
echo "Host: $RABBITMQ_HOST"
echo "Port: $RABBITMQ_PORT (Management), 5672 (AMQP)"
echo "Virtual Host: $RABBITMQ_VHOST"
echo "Admin User: $RABBITMQ_USER"
echo "App User: app_user"
echo -e "\n${YELLOW}Management UI:${NC} http://$RABBITMQ_HOST:$RABBITMQ_PORT"

# Display queue information
echo -e "\n${YELLOW}Available Queues:${NC}"
echo "- video.processing.queue (TTL: 1h, Max Retries: 3)"
echo "- video.conversion.queue (TTL: 2h, Max Retries: 2)"
echo "- analytics.processing.queue (TTL: 30m, Max Retries: 3)"
echo "- notification.email.queue (TTL: 10m, Max Retries: 5)"
echo "- dlx.failed.queue (Dead Letter Queue)"