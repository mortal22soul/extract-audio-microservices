#!/bin/bash

set -e

echo "🚀 Setting up development environment..."

# Check if required tools are installed
check_tool() {
    if ! command -v $1 &> /dev/null; then
        echo "❌ $1 is not installed. Please install it first."
        exit 1
    else
        echo "✅ $1 is installed"
    fi
}

echo "📋 Checking required tools..."
check_tool "go"
check_tool "node"
check_tool "pnpm"
check_tool "python3"
check_tool "uv"
check_tool "docker"
check_tool "docker-compose"

# Check optional tools
echo "📋 Checking optional tools..."
if command -v tilt &> /dev/null; then
    echo "✅ tilt is installed"
    TILT_AVAILABLE=true
else
    echo "⚠️  tilt is not installed (optional for Kubernetes development)"
    TILT_AVAILABLE=false
fi

if command -v kubectl &> /dev/null; then
    echo "✅ kubectl is installed"
    KUBECTL_AVAILABLE=true
else
    echo "⚠️  kubectl is not installed (optional for Kubernetes development)"
    KUBECTL_AVAILABLE=false
fi

# Install Go tools
echo "🔧 Installing Go development tools..."
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
go install golang.org/x/tools/cmd/goimports@latest

# Install Python tools globally
echo "🐍 Installing Python development tools..."
pip install pre-commit

# Install dependencies
echo "📦 Installing project dependencies..."
make deps

# Set up pre-commit hooks
echo "🪝 Setting up pre-commit hooks..."
pre-commit install
pre-commit install --hook-type commit-msg

# Generate protocol buffers
echo "🔄 Generating protocol buffers..."
if command -v buf &> /dev/null; then
    make proto
else
    echo "⚠️  buf is not installed, skipping protocol buffer generation"
    echo "   Install buf from https://buf.build/docs/installation"
fi

echo ""
echo "🎉 Development environment setup complete!"
echo ""
echo "📚 Available commands:"
echo "  make help           - Show all available commands"
echo "  make dev            - Start all services with Tilt (requires Kubernetes)"
echo "  make docker-up      - Start all services with Docker Compose"
echo "  make test           - Run all tests"
echo "  make lint           - Run all linters"
echo "  make format         - Format all code"
echo "  make check          - Run all quality checks"
echo ""

if [ "$TILT_AVAILABLE" = true ] && [ "$KUBECTL_AVAILABLE" = true ]; then
    echo "🚀 Kubernetes development ready! Run 'make dev' to start with Tilt"
else
    echo "🐳 Docker development ready! Run 'make docker-up' to start with Docker Compose"
fi

echo ""
echo "💡 Pro tips:"
echo "  - Run 'make pre-commit-run' to check all files before committing"
echo "  - Use 'make lint-fix' to auto-fix common linting issues"
echo "  - Check 'make help' for all available commands"