#!/bin/bash

# PostgreSQL migration script using golang-migrate
# Usage: ./migrate.sh [up|down|version|force] [steps]

set -e

# Configuration
DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-5432}
DB_NAME=${DB_NAME:-video_converter_auth}
DB_USER=${DB_USER:-app_user}
DB_PASSWORD=${DB_PASSWORD:-dev_password_123}
MIGRATIONS_PATH="file://$(pwd)/migrations"

# Database URL
DATABASE_URL="postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=disable"

# Check if migrate is installed
if ! command -v migrate &> /dev/null; then
    echo "golang-migrate is not installed. Installing..."
    
    # Install migrate based on OS
    if [[ "$OSTYPE" == "linux-gnu"* ]]; then
        curl -L https://github.com/golang-migrate/migrate/releases/latest/download/migrate.linux-amd64.tar.gz | tar xvz
        sudo mv migrate /usr/local/bin/
    elif [[ "$OSTYPE" == "darwin"* ]]; then
        brew install golang-migrate
    elif [[ "$OSTYPE" == "msys" ]] || [[ "$OSTYPE" == "win32" ]]; then
        echo "Please install golang-migrate manually from: https://github.com/golang-migrate/migrate/releases"
        exit 1
    fi
fi

# Function to run migrations
run_migration() {
    local command=$1
    local steps=$2
    
    echo "Running migration: $command"
    echo "Database URL: postgres://${DB_USER}:***@${DB_HOST}:${DB_PORT}/${DB_NAME}"
    
    case $command in
        "up")
            if [ -n "$steps" ]; then
                migrate -path migrations -database "$DATABASE_URL" up "$steps"
            else
                migrate -path migrations -database "$DATABASE_URL" up
            fi
            ;;
        "down")
            if [ -n "$steps" ]; then
                migrate -path migrations -database "$DATABASE_URL" down "$steps"
            else
                echo "Warning: This will drop all tables. Use 'down 1' to rollback one migration."
                read -p "Are you sure? (y/N): " -n 1 -r
                echo
                if [[ $REPLY =~ ^[Yy]$ ]]; then
                    migrate -path migrations -database "$DATABASE_URL" down
                fi
            fi
            ;;
        "version")
            migrate -path migrations -database "$DATABASE_URL" version
            ;;
        "force")
            if [ -z "$steps" ]; then
                echo "Error: force command requires a version number"
                exit 1
            fi
            migrate -path migrations -database "$DATABASE_URL" force "$steps"
            ;;
        "create")
            if [ -z "$steps" ]; then
                echo "Error: create command requires a migration name"
                exit 1
            fi
            migrate create -ext sql -dir migrations -seq "$steps"
            ;;
        *)
            echo "Usage: $0 [up|down|version|force|create] [steps/version/name]"
            echo "Examples:"
            echo "  $0 up           # Run all pending migrations"
            echo "  $0 up 1         # Run next migration"
            echo "  $0 down 1       # Rollback last migration"
            echo "  $0 version      # Show current version"
            echo "  $0 force 1      # Force version (use with caution)"
            echo "  $0 create add_user_preferences  # Create new migration"
            exit 1
            ;;
    esac
}

# Main execution
run_migration "$1" "$2"