@echo off
REM Redis setup script for Windows
REM This script initializes Redis with the required data structures and configurations

setlocal enabledelayedexpansion

REM Configuration
if "%REDIS_HOST%"=="" set REDIS_HOST=localhost
if "%REDIS_PORT%"=="" set REDIS_PORT=6379
if "%REDIS_PASSWORD%"=="" set REDIS_PASSWORD=dev_redis_password_123

echo Setting up Redis for video converter microservices...

REM Check if redis-cli is available
redis-cli --version >nul 2>&1
if errorlevel 1 (
    echo redis-cli is not installed. Please install Redis client tools.
    exit /b 1
)

REM Test Redis connection
echo Testing Redis connection...
redis-cli -h %REDIS_HOST% -p %REDIS_PORT% -a %REDIS_PASSWORD% ping >nul 2>&1
if errorlevel 1 (
    echo Cannot connect to Redis at %REDIS_HOST%:%REDIS_PORT%
    echo Please ensure Redis is running and the credentials are correct.
    exit /b 1
)

echo Redis connection successful!

REM Set up pub/sub channels
echo Setting up pub/sub channels...
redis-cli -h %REDIS_HOST% -p %REDIS_PORT% -a %REDIS_PASSWORD% HSET channels:info conversion:progress "Video conversion progress updates"
redis-cli -h %REDIS_HOST% -p %REDIS_PORT% -a %REDIS_PASSWORD% HSET channels:info conversion:complete "Video conversion completion notifications"
redis-cli -h %REDIS_HOST% -p %REDIS_PORT% -a %REDIS_PASSWORD% HSET channels:info conversion:error "Video conversion error notifications"
redis-cli -h %REDIS_HOST% -p %REDIS_PORT% -a %REDIS_PASSWORD% HSET channels:info user:activity "User activity tracking"
redis-cli -h %REDIS_HOST% -p %REDIS_PORT% -a %REDIS_PASSWORD% HSET channels:info system:alerts "System-wide alerts and notifications"

REM Set up configuration
echo Setting up Redis data structures...
redis-cli -h %REDIS_HOST% -p %REDIS_PORT% -a %REDIS_PASSWORD% HSET config:app max_file_size 104857600
redis-cli -h %REDIS_HOST% -p %REDIS_PORT% -a %REDIS_PASSWORD% HSET config:app supported_formats "mp4,avi,mov,mkv,webm,flv"
redis-cli -h %REDIS_HOST% -p %REDIS_PORT% -a %REDIS_PASSWORD% HSET config:app max_conversion_time 3600
redis-cli -h %REDIS_HOST% -p %REDIS_PORT% -a %REDIS_PASSWORD% HSET config:app max_concurrent_jobs 5
redis-cli -h %REDIS_HOST% -p %REDIS_PORT% -a %REDIS_PASSWORD% HSET config:app cleanup_interval 86400

REM Rate limiting configurations
redis-cli -h %REDIS_HOST% -p %REDIS_PORT% -a %REDIS_PASSWORD% HSET config:rate_limits upload_per_hour 10
redis-cli -h %REDIS_HOST% -p %REDIS_PORT% -a %REDIS_PASSWORD% HSET config:rate_limits api_requests_per_minute 100
redis-cli -h %REDIS_HOST% -p %REDIS_PORT% -a %REDIS_PASSWORD% HSET config:rate_limits conversion_per_day 50

REM Cache TTL settings
redis-cli -h %REDIS_HOST% -p %REDIS_PORT% -a %REDIS_PASSWORD% HSET config:ttl user_session 86400
redis-cli -h %REDIS_HOST% -p %REDIS_PORT% -a %REDIS_PASSWORD% HSET config:ttl conversion_progress 3600
redis-cli -h %REDIS_HOST% -p %REDIS_PORT% -a %REDIS_PASSWORD% HSET config:ttl user_activity 604800
redis-cli -h %REDIS_HOST% -p %REDIS_PORT% -a %REDIS_PASSWORD% HSET config:ttl rate_limit 3600

REM System status
for /f %%i in ('powershell -command "[int][double]::Parse((Get-Date -UFormat %%s))"') do set timestamp=%%i
redis-cli -h %REDIS_HOST% -p %REDIS_PORT% -a %REDIS_PASSWORD% HSET system:status redis_initialized 1
redis-cli -h %REDIS_HOST% -p %REDIS_PORT% -a %REDIS_PASSWORD% HSET system:status last_setup_time %timestamp%
redis-cli -h %REDIS_HOST% -p %REDIS_PORT% -a %REDIS_PASSWORD% HSET system:status version "1.0.0"

echo Setting up sorted sets for analytics...
redis-cli -h %REDIS_HOST% -p %REDIS_PORT% -a %REDIS_PASSWORD% ZADD stats:conversions:daily %timestamp% 0
redis-cli -h %REDIS_HOST% -p %REDIS_PORT% -a %REDIS_PASSWORD% ZADD stats:uploads:daily %timestamp% 0
redis-cli -h %REDIS_HOST% -p %REDIS_PORT% -a %REDIS_PASSWORD% ZADD stats:users:active %timestamp% 0

REM Popular video formats
redis-cli -h %REDIS_HOST% -p %REDIS_PORT% -a %REDIS_PASSWORD% ZADD stats:formats 0 mp4
redis-cli -h %REDIS_HOST% -p %REDIS_PORT% -a %REDIS_PASSWORD% ZADD stats:formats 0 avi
redis-cli -h %REDIS_HOST% -p %REDIS_PORT% -a %REDIS_PASSWORD% ZADD stats:formats 0 mov
redis-cli -h %REDIS_HOST% -p %REDIS_PORT% -a %REDIS_PASSWORD% ZADD stats:formats 0 mkv
redis-cli -h %REDIS_HOST% -p %REDIS_PORT% -a %REDIS_PASSWORD% ZADD stats:formats 0 webm

echo Configuring connection pool settings...
redis-cli -h %REDIS_HOST% -p %REDIS_PORT% -a %REDIS_PASSWORD% HSET pool:auth max_connections 20
redis-cli -h %REDIS_HOST% -p %REDIS_PORT% -a %REDIS_PASSWORD% HSET pool:auth min_connections 5
redis-cli -h %REDIS_HOST% -p %REDIS_PORT% -a %REDIS_PASSWORD% HSET pool:auth connection_timeout 5000

redis-cli -h %REDIS_HOST% -p %REDIS_PORT% -a %REDIS_PASSWORD% HSET pool:gateway max_connections 50
redis-cli -h %REDIS_HOST% -p %REDIS_PORT% -a %REDIS_PASSWORD% HSET pool:gateway min_connections 10
redis-cli -h %REDIS_HOST% -p %REDIS_PORT% -a %REDIS_PASSWORD% HSET pool:gateway connection_timeout 5000

redis-cli -h %REDIS_HOST% -p %REDIS_PORT% -a %REDIS_PASSWORD% HSET pool:converter max_connections 30
redis-cli -h %REDIS_HOST% -p %REDIS_PORT% -a %REDIS_PASSWORD% HSET pool:converter min_connections 5
redis-cli -h %REDIS_HOST% -p %REDIS_PORT% -a %REDIS_PASSWORD% HSET pool:converter connection_timeout 10000

redis-cli -h %REDIS_HOST% -p %REDIS_PORT% -a %REDIS_PASSWORD% HSET pool:realtime max_connections 100
redis-cli -h %REDIS_HOST% -p %REDIS_PORT% -a %REDIS_PASSWORD% HSET pool:realtime min_connections 20
redis-cli -h %REDIS_HOST% -p %REDIS_PORT% -a %REDIS_PASSWORD% HSET pool:realtime connection_timeout 3000

echo Verifying Redis setup...
redis-cli -h %REDIS_HOST% -p %REDIS_PORT% -a %REDIS_PASSWORD% EXISTS config:app
redis-cli -h %REDIS_HOST% -p %REDIS_PORT% -a %REDIS_PASSWORD% HGETALL system:status

echo.
echo Redis setup completed successfully!
echo Pub/sub channels configured for real-time communication
echo Data structures initialized for caching and session management
echo Rate limiting and connection pooling configured
echo.
echo Redis Connection Information:
echo Host: %REDIS_HOST%
echo Port: %REDIS_PORT%
echo Password: [CONFIGURED]
echo.
echo Available Pub/Sub Channels:
echo - conversion:progress
echo - conversion:complete
echo - conversion:error
echo - user:activity
echo - system:alerts