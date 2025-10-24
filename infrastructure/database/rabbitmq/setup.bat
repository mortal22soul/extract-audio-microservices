@echo off
REM RabbitMQ setup script for Windows
REM This script configures exchanges, queues, bindings, and policies

setlocal enabledelayedexpansion

REM Configuration
if "%RABBITMQ_HOST%"=="" set RABBITMQ_HOST=localhost
if "%RABBITMQ_PORT%"=="" set RABBITMQ_PORT=15672
if "%RABBITMQ_USER%"=="" set RABBITMQ_USER=admin
if "%RABBITMQ_PASSWORD%"=="" set RABBITMQ_PASSWORD=admin_password_123
if "%RABBITMQ_VHOST%"=="" set RABBITMQ_VHOST=video_converter

echo Setting up RabbitMQ for video converter microservices...

REM Check if rabbitmqadmin is available
where rabbitmqadmin >nul 2>&1
if errorlevel 1 (
    echo rabbitmqadmin not found. Please download it from:
    echo http://%RABBITMQ_HOST%:%RABBITMQ_PORT%/cli/rabbitmqadmin
    echo And place it in your PATH or current directory.
    exit /b 1
)

REM Base command with authentication
set RABBITMQ_CMD=rabbitmqadmin -H %RABBITMQ_HOST% -P %RABBITMQ_PORT% -u %RABBITMQ_USER% -p %RABBITMQ_PASSWORD%

REM Test RabbitMQ connection
echo Testing RabbitMQ connection...
%RABBITMQ_CMD% list vhosts >nul 2>&1
if errorlevel 1 (
    echo Cannot connect to RabbitMQ at %RABBITMQ_HOST%:%RABBITMQ_PORT%
    echo Please ensure RabbitMQ is running and the credentials are correct.
    exit /b 1
)

echo RabbitMQ connection successful!

REM Create virtual host if it doesn't exist
echo Creating virtual host: %RABBITMQ_VHOST%
%RABBITMQ_CMD% declare vhost name=%RABBITMQ_VHOST% 2>nul || echo Virtual host may already exist

REM Set permissions for the virtual host
echo Setting permissions for virtual host...
rabbitmqctl set_permissions -p %RABBITMQ_VHOST% %RABBITMQ_USER% ".*" ".*" ".*" 2>nul || echo Permissions may already be set

REM Create exchanges
echo Creating exchanges...
%RABBITMQ_CMD% -V %RABBITMQ_VHOST% declare exchange name=video.exchange type=topic durable=true
%RABBITMQ_CMD% -V %RABBITMQ_VHOST% declare exchange name=analytics.exchange type=topic durable=true
%RABBITMQ_CMD% -V %RABBITMQ_VHOST% declare exchange name=notification.exchange type=direct durable=true
%RABBITMQ_CMD% -V %RABBITMQ_VHOST% declare exchange name=dlx.exchange type=direct durable=true

REM Create queues with dead letter exchange configuration
echo Creating queues...

REM Video processing queue
%RABBITMQ_CMD% -V %RABBITMQ_VHOST% declare queue name=video.processing.queue durable=true arguments="{\"x-message-ttl\":3600000,\"x-dead-letter-exchange\":\"dlx.exchange\",\"x-dead-letter-routing-key\":\"video.failed\",\"x-max-retries\":3}"

REM Video conversion queue
%RABBITMQ_CMD% -V %RABBITMQ_VHOST% declare queue name=video.conversion.queue durable=true arguments="{\"x-message-ttl\":7200000,\"x-dead-letter-exchange\":\"dlx.exchange\",\"x-dead-letter-routing-key\":\"conversion.failed\",\"x-max-retries\":2}"

REM Analytics processing queue
%RABBITMQ_CMD% -V %RABBITMQ_VHOST% declare queue name=analytics.processing.queue durable=true arguments="{\"x-message-ttl\":1800000,\"x-dead-letter-exchange\":\"dlx.exchange\",\"x-dead-letter-routing-key\":\"analytics.failed\",\"x-max-retries\":3}"

REM Notification email queue
%RABBITMQ_CMD% -V %RABBITMQ_VHOST% declare queue name=notification.email.queue durable=true arguments="{\"x-message-ttl\":600000,\"x-dead-letter-exchange\":\"dlx.exchange\",\"x-dead-letter-routing-key\":\"notification.failed\",\"x-max-retries\":5}"

REM Dead letter queue
%RABBITMQ_CMD% -V %RABBITMQ_VHOST% declare queue name=dlx.failed.queue durable=true

REM Create bindings
echo Creating bindings...

REM Video exchange bindings
%RABBITMQ_CMD% -V %RABBITMQ_VHOST% declare binding source=video.exchange destination=video.processing.queue routing_key=video.uploaded
%RABBITMQ_CMD% -V %RABBITMQ_VHOST% declare binding source=video.exchange destination=video.conversion.queue routing_key=video.convert
%RABBITMQ_CMD% -V %RABBITMQ_VHOST% declare binding source=video.exchange destination=analytics.processing.queue routing_key=video.analyze

REM Analytics exchange bindings
%RABBITMQ_CMD% -V %RABBITMQ_VHOST% declare binding source=analytics.exchange destination=analytics.processing.queue routing_key=analytics.*

REM Notification exchange bindings
%RABBITMQ_CMD% -V %RABBITMQ_VHOST% declare binding source=notification.exchange destination=notification.email.queue routing_key=email

REM Dead letter exchange bindings
%RABBITMQ_CMD% -V %RABBITMQ_VHOST% declare binding source=dlx.exchange destination=dlx.failed.queue routing_key=failed

REM Set up policies
echo Setting up policies...

REM High availability policy
rabbitmqctl set_policy -p %RABBITMQ_VHOST% ha-all ".*" "{\"ha-mode\":\"all\",\"ha-sync-mode\":\"automatic\"}" --priority 0 2>nul || echo HA policy may already exist

REM Dead letter exchange policy for specific queues
rabbitmqctl set_policy -p %RABBITMQ_VHOST% dlx-policy "^(video|notification|analytics)\." "{\"dead-letter-exchange\":\"dlx.exchange\",\"dead-letter-routing-key\":\"failed\",\"message-ttl\":3600000}" --priority 1 2>nul || echo DLX policy may already exist

REM Create application user
echo Creating application user...
rabbitmqctl add_user app_user app_password_123 2>nul || echo User may already exist
rabbitmqctl set_user_tags app_user management 2>nul || echo User tags may already be set
rabbitmqctl set_permissions -p %RABBITMQ_VHOST% app_user ".*" ".*" ".*" 2>nul || echo User permissions may already be set

REM Verify setup
echo Verifying RabbitMQ setup...
echo Exchanges:
%RABBITMQ_CMD% -V %RABBITMQ_VHOST% list exchanges name type durable

echo.
echo Queues:
%RABBITMQ_CMD% -V %RABBITMQ_VHOST% list queues name messages durable

echo.
echo Bindings:
%RABBITMQ_CMD% -V %RABBITMQ_VHOST% list bindings source_name destination_name routing_key

echo.
echo RabbitMQ setup completed successfully!
echo Exchanges, queues, and bindings configured
echo Dead letter queues set up for error handling
echo High availability and retry policies applied

echo.
echo RabbitMQ Connection Information:
echo Host: %RABBITMQ_HOST%
echo Port: %RABBITMQ_PORT% (Management), 5672 (AMQP)
echo Virtual Host: %RABBITMQ_VHOST%
echo Admin User: %RABBITMQ_USER%
echo App User: app_user
echo.
echo Management UI: http://%RABBITMQ_HOST%:%RABBITMQ_PORT%

echo.
echo Available Queues:
echo - video.processing.queue (TTL: 1h, Max Retries: 3)
echo - video.conversion.queue (TTL: 2h, Max Retries: 2)
echo - analytics.processing.queue (TTL: 30m, Max Retries: 3)
echo - notification.email.queue (TTL: 10m, Max Retries: 5)
echo - dlx.failed.queue (Dead Letter Queue)