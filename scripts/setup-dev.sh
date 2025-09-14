#!/bin/bash

# MSC Shard Router - Development Setup Script
# This script sets up the complete development environment

set -e

echo "🚀 MSC Shard Router - Development Setup"
echo "======================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Check Prerequisites
print_status "Checking prerequisites..."

# Check Go
if command_exists go; then
    GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    print_success "Go $GO_VERSION found"
    
    # Check Go version (require 1.21+)
    if [[ "$(printf '%s\n' "1.21" "$GO_VERSION" | sort -V | head -n1)" = "1.21" ]]; then
        print_success "Go version is compatible (>= 1.21)"
    else
        print_error "Go version $GO_VERSION is too old. Please install Go 1.21 or later"
        exit 1
    fi
else
    print_error "Go is not installed. Please install Go 1.21 or later"
    echo "Visit: https://golang.org/dl/"
    exit 1
fi

# Check Docker
if command_exists docker; then
    print_success "Docker found"
else
    print_warning "Docker not found. Docker is recommended for local development"
    echo "Visit: https://docs.docker.com/get-docker/"
fi

# Check Docker Compose
if command_exists docker-compose || docker compose version >/dev/null 2>&1; then
    print_success "Docker Compose found"
else
    print_warning "Docker Compose not found. Required for integration tests"
fi

# Check Make
if command_exists make; then
    print_success "Make found"
else
    print_error "Make is required for build automation"
    exit 1
fi

# Check Git
if command_exists git; then
    print_success "Git found"
else
    print_error "Git is required for version control"
    exit 1
fi

echo ""

# Setup Go Module
print_status "Setting up Go module..."
if [ -f "go.mod" ]; then
    go mod download
    go mod tidy
    print_success "Go dependencies downloaded and tidied"
else
    print_error "go.mod not found. Are you in the correct directory?"
    exit 1
fi

# Install development tools
print_status "Installing development tools..."

# golangci-lint
if ! command_exists golangci-lint; then
    print_status "Installing golangci-lint..."
    curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.55.2
    if command_exists golangci-lint; then
        print_success "golangci-lint installed"
    else
        print_warning "Failed to install golangci-lint. Install manually for linting"
    fi
else
    print_success "golangci-lint already installed"
fi

# gosec
if ! command_exists gosec; then
    print_status "Installing gosec..."
    go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
    if command_exists gosec; then
        print_success "gosec installed"
    else
        print_warning "Failed to install gosec. Install manually for security scanning"
    fi
else
    print_success "gosec already installed"
fi

# goimports
if ! command_exists goimports; then
    print_status "Installing goimports..."
    go install golang.org/x/tools/cmd/goimports@latest
    if command_exists goimports; then
        print_success "goimports installed"
    else
        print_warning "Failed to install goimports. Install manually for import formatting"
    fi
else
    print_success "goimports already installed"
fi

echo ""

# Validate setup
print_status "Validating setup..."

# Build project
print_status "Building project..."
if make build >/dev/null 2>&1; then
    print_success "Build successful"
else
    print_error "Build failed"
    exit 1
fi

# Run tests
print_status "Running tests..."
if make test >/dev/null 2>&1; then
    print_success "All tests passed"
else
    print_error "Tests failed"
    exit 1
fi

# Run linting
print_status "Running linting..."
if make lint >/dev/null 2>&1; then
    print_success "Linting passed"
else
    print_warning "Linting issues found. Run 'make lint' for details"
fi

echo ""

# Setup Git hooks (optional)
if [ -d ".git" ]; then
    print_status "Setting up Git hooks..."
    
    # Pre-commit hook
    cat > .git/hooks/pre-commit << 'EOF'
#!/bin/bash
echo "Running pre-commit checks..."

# Run tests
if ! make test; then
    echo "Tests failed. Commit aborted."
    exit 1
fi

# Run linting
if ! make lint; then
    echo "Linting failed. Commit aborted."
    exit 1
fi

# Run security scan
if ! make security; then
    echo "Security scan failed. Commit aborted."
    exit 1
fi

echo "Pre-commit checks passed!"
EOF
    
    chmod +x .git/hooks/pre-commit
    print_success "Git pre-commit hook installed"
else
    print_warning "Not a Git repository. Skipping Git hooks setup"
fi

echo ""

# Setup environment file
print_status "Creating development environment file..."
cat > .env.development << 'EOF'
# MSC Shard Router - Development Environment
# Copy to .env and modify as needed

# Router Configuration
ROUTER_PORT=8080
SHARDING_KEY=id_client

# Shard URLs (add more as needed)
SHARD_01_URL=http://localhost:8081
SHARD_02_URL=http://localhost:8082
SHARD_03_URL=http://localhost:8083
SHARD_04_URL=http://localhost:8084

# Development Settings
LOG_LEVEL=debug
METRICS_PORT=9090
HEALTH_CHECK_INTERVAL=30s

# Hash Ring Configuration
VIRTUAL_REPLICAS=3
HASH_ALGORITHM=sha512
EOF

print_success "Development environment file created (.env.development)"

echo ""

# Final summary
print_success "🎉 Development environment setup complete!"
echo ""
echo "📋 Next steps:"
echo "  1. Copy .env.development to .env and modify as needed"
echo "  2. Start development environment: make docker-compose-up"
echo "  3. Run the application: make run"
echo "  4. Run tests: make test"
echo "  5. See available commands: make help"
echo ""
echo "📚 Useful commands:"
echo "  make test-coverage    # Run tests with coverage report"
echo "  make lint            # Run linting checks"
echo "  make security        # Run security scans"
echo "  make benchmark       # Run performance benchmarks"
echo "  make ci-full         # Run complete CI pipeline locally"
echo ""
echo "🔗 Resources:"
echo "  - README.md          # Project documentation"
echo "  - CONTRIBUTING.md    # Contribution guidelines"
echo "  - CHANGELOG.md       # Version history"
echo ""
print_success "Happy coding! 🚀"