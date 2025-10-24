@echo off
REM Complete database setup script for Windows
REM This script sets up PostgreSQL, MongoDB, Redis, and RabbitMQ

setlocal enabledelayedexpansion

echo ========================================
echo Video Converter Database Setup
echo ========================================

REM Get script directory
set SCRIPT_DIR=%~dp0
set DOCKER_DIR=%SCRIPT_DIR%docker
set DATABASE_DIR=%SCRIPT_DIR%database

REM Check if Docker and Docker Compose are installed
docker --version >nul 2>&1
if errorlevel 1 (
    echo Docker is not installed. Please install Docker first.
    exit /b 1
)

docker-compose --version >nul 2>&1
if errorlevel 1 (
    docker compose version >nul 2>&1
    if errorlevel 1 (
        echo Docker Compose is not installed. Please install Docker Compose first.
        exit /b 1
    ) else (
        set DOCKER_COMPOSE_CMD=docker compose
    )
) else (
    set DOCKER_COMPOSE_CMD=docker-compose
)

echo ✓ Docker and Docker Compose are available

REM Start database services
echo Starting database services...
cd /d "%SCRIPT_DIR%"
%DOCKER_COMPOSE_CMD% -f docker/databases.yml up -d

REM Wait for services to start
echo Waiting for services to start...
timeout /t 30 /nobreak >nul

REM Check service status
echo Checking service status...
%DOCKER_COMPOSE_CMD% -f docker/databases.yml ps

REM Run database-specific setup scripts
echo Running database setup scripts...

REM PostgreSQL migrations
if exist "%DATABASE_DIR%\postgresql\migrate.bat" (
    echo Running PostgreSQL migrations...
    cd /d "%DATABASE_DIR%\postgresql"
    call migrate.bat up 2>nul || echo PostgreSQL migrations may have already been applied
)

REM Redis setup
if exist "%DATABASE_DIR%\redis\setup.bat" (
    echo Setting up Redis...
    cd /d "%DATABASE_DIR%\redis"
    call setup.bat 2>nul || echo Redis setup may have already been completed
)

REM RabbitMQ setup
if exist "%DATABASE_DIR%\rabbitmq\setup.bat" (
    echo Setting up RabbitMQ...
    cd /d "%DATABASE_DIR%\rabbitmq"
    call setup.bat 2>nul || echo RabbitMQ setup may have already been completed
)

REM Verify all services are running
echo Verifying service status...
cd /d "%SCRIPT_DIR%"
%DOCKER_COMPOSE_CMD% -f docker/databases.yml ps

echo ========================================
echo Database Setup Complete!
echo ========================================

echo Services are running on the following ports:
echo PostgreSQL:     localhost:5432
echo MongoDB:        localhost:27017
echo Redis:          localhost:6379
echo RabbitMQ AMQP:  localhost:5672
echo RabbitMQ Mgmt:  localhost:15672

echo.
echo Admin interfaces (use --profile admin-tools):
echo pgAdmin:        http://localhost:8080 (admin@example.com / admin123)
echo Mongo Express:  http://localhost:8081 (admin / admin123)
echo Redis Commander:http://localhost:8082 (admin / admin123)
echo RabbitMQ Mgmt:  http://localhost:15672 (admin / admin_password_123)

echo.
echo Database credentials:
echo PostgreSQL: app_user / dev_password_123
echo MongoDB:    app_user / dev_password_123
echo Redis:      password: dev_redis_password_123
echo RabbitMQ:   app_user / app_password_123

echo.
echo To start admin tools:
echo %DOCKER_COMPOSE_CMD% -f docker/databases.yml --profile admin-tools up -d

echo.
echo To stop all services:
echo %DOCKER_COMPOSE_CMD% -f docker/databases.yml down

echo.
echo To view logs:
echo %DOCKER_COMPOSE_CMD% -f docker/databases.yml logs -f [service_name]