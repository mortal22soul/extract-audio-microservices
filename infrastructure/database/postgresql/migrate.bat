@echo off
REM PostgreSQL migration script for Windows using golang-migrate
REM Usage: migrate.bat [up|down|version|force] [steps]

setlocal enabledelayedexpansion

REM Configuration
if "%DB_HOST%"=="" set DB_HOST=localhost
if "%DB_PORT%"=="" set DB_PORT=5432
if "%DB_NAME%"=="" set DB_NAME=video_converter_auth
if "%DB_USER%"=="" set DB_USER=app_user
if "%DB_PASSWORD%"=="" set DB_PASSWORD=dev_password_123

REM Database URL
set DATABASE_URL=postgres://%DB_USER%:%DB_PASSWORD%@%DB_HOST%:%DB_PORT%/%DB_NAME%?sslmode=disable

REM Check if migrate is installed
migrate version >nul 2>&1
if errorlevel 1 (
    echo golang-migrate is not installed. Please install it from:
    echo https://github.com/golang-migrate/migrate/releases
    exit /b 1
)

set COMMAND=%1
set STEPS=%2

if "%COMMAND%"=="" (
    goto :usage
)

echo Running migration: %COMMAND%
echo Database URL: postgres://%DB_USER%:***@%DB_HOST%:%DB_PORT%/%DB_NAME%

if "%COMMAND%"=="up" (
    if "%STEPS%"=="" (
        migrate -path migrations -database "%DATABASE_URL%" up
    ) else (
        migrate -path migrations -database "%DATABASE_URL%" up %STEPS%
    )
) else if "%COMMAND%"=="down" (
    if "%STEPS%"=="" (
        echo Warning: This will drop all tables. Use 'down 1' to rollback one migration.
        set /p CONFIRM="Are you sure? (y/N): "
        if /i "!CONFIRM!"=="y" (
            migrate -path migrations -database "%DATABASE_URL%" down
        )
    ) else (
        migrate -path migrations -database "%DATABASE_URL%" down %STEPS%
    )
) else if "%COMMAND%"=="version" (
    migrate -path migrations -database "%DATABASE_URL%" version
) else if "%COMMAND%"=="force" (
    if "%STEPS%"=="" (
        echo Error: force command requires a version number
        exit /b 1
    )
    migrate -path migrations -database "%DATABASE_URL%" force %STEPS%
) else if "%COMMAND%"=="create" (
    if "%STEPS%"=="" (
        echo Error: create command requires a migration name
        exit /b 1
    )
    migrate create -ext sql -dir migrations -seq %STEPS%
) else (
    goto :usage
)

goto :end

:usage
echo Usage: %0 [up^|down^|version^|force^|create] [steps/version/name]
echo Examples:
echo   %0 up           # Run all pending migrations
echo   %0 up 1         # Run next migration
echo   %0 down 1       # Rollback last migration
echo   %0 version      # Show current version
echo   %0 force 1      # Force version (use with caution)
echo   %0 create add_user_preferences  # Create new migration
exit /b 1

:end