@echo off
setlocal enabledelayedexpansion

echo 🚀 Setting up development environment...

:: Check if required tools are installed
echo 📋 Checking required tools...

where go >nul 2>nul
if %errorlevel% neq 0 (
    echo ❌ go is not installed. Please install it first.
    exit /b 1
) else (
    echo ✅ go is installed
)

where node >nul 2>nul
if %errorlevel% neq 0 (
    echo ❌ node is not installed. Please install it first.
    exit /b 1
) else (
    echo ✅ node is installed
)

where pnpm >nul 2>nul
if %errorlevel% neq 0 (
    echo ❌ pnpm is not installed. Please install it first.
    exit /b 1
) else (
    echo ✅ pnpm is installed
)

where python >nul 2>nul
if %errorlevel% neq 0 (
    echo ❌ python is not installed. Please install it first.
    exit /b 1
) else (
    echo ✅ python is installed
)

where uv >nul 2>nul
if %errorlevel% neq 0 (
    echo ❌ uv is not installed. Please install it first.
    exit /b 1
) else (
    echo ✅ uv is installed
)

where docker >nul 2>nul
if %errorlevel% neq 0 (
    echo ❌ docker is not installed. Please install it first.
    exit /b 1
) else (
    echo ✅ docker is installed
)

:: Check optional tools
echo 📋 Checking optional tools...
where tilt >nul 2>nul
if %errorlevel% neq 0 (
    echo ⚠️  tilt is not installed (optional for Kubernetes development)
    set TILT_AVAILABLE=false
) else (
    echo ✅ tilt is installed
    set TILT_AVAILABLE=true
)

where kubectl >nul 2>nul
if %errorlevel% neq 0 (
    echo ⚠️  kubectl is not installed (optional for Kubernetes development)
    set KUBECTL_AVAILABLE=false
) else (
    echo ✅ kubectl is installed
    set KUBECTL_AVAILABLE=true
)

:: Install Go tools
echo 🔧 Installing Go development tools...
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install golang.org/x/tools/cmd/goimports@latest

:: Install Python tools globally
echo 🐍 Installing Python development tools...
pip install pre-commit

:: Install dependencies
echo 📦 Installing project dependencies...
make deps

:: Set up pre-commit hooks
echo 🪝 Setting up pre-commit hooks...
pre-commit install
pre-commit install --hook-type commit-msg

:: Generate protocol buffers
echo 🔄 Generating protocol buffers...
where buf >nul 2>nul
if %errorlevel% neq 0 (
    echo ⚠️  buf is not installed, skipping protocol buffer generation
    echo    Install buf from https://buf.build/docs/installation
) else (
    make proto
)

echo.
echo 🎉 Development environment setup complete!
echo.
echo 📚 Available commands:
echo   make help           - Show all available commands
echo   make dev            - Start all services with Tilt (requires Kubernetes)
echo   make docker-up      - Start all services with Docker Compose
echo   make test           - Run all tests
echo   make lint           - Run all linters
echo   make format         - Format all code
echo   make check          - Run all quality checks
echo.

if "!TILT_AVAILABLE!"=="true" if "!KUBECTL_AVAILABLE!"=="true" (
    echo 🚀 Kubernetes development ready! Run 'make dev' to start with Tilt
) else (
    echo 🐳 Docker development ready! Run 'make docker-up' to start with Docker Compose
)

echo.
echo 💡 Pro tips:
echo   - Run 'make pre-commit-run' to check all files before committing
echo   - Use 'make lint-fix' to auto-fix common linting issues
echo   - Check 'make help' for all available commands