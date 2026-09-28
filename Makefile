.PHONY: help build run test docker docker-run docker-build clean dev install

# Default target
help:
	@echo "Available targets:"
	@echo "  make build       - Build the binary"
	@echo "  make run         - Run the server"
	@echo "  make dev         - Run with hot reload (requires air)"
	@echo "  make test        - Run tests"
	@echo "  make docker      - Build and run Docker container"
	@echo "  make docker-build - Build Docker image"
	@echo "  make docker-run  - Run Docker container"
	@echo "  make install     - Install dependencies"
	@echo "  make clean       - Clean build artifacts"

# Build the application
build: fmt
	@echo "🔨 Building..."
	go build -o bin/cors-proxy main.go
	@echo "✅ Build complete: bin/cors-proxy"

# Run the application
run:
	@echo "🚀 Starting server..."
	go run main.go

# Development mode with hot reload (requires air)
dev:
	@echo "🔥 Starting development server with hot reload..."
	@command -v air > /dev/null 2>&1 || (echo "Installing air..." && go install github.com/cosmtrek/air@latest)
	air

# Run tests
test:
	@echo "🧪 Running tests..."
	chmod +x test.sh
	./test.sh

# Install dependencies
install:
	@echo "📦 Installing dependencies..."
	go mod download
	go mod tidy

# Build Docker image
docker-build:
	@echo "🐳 Building Docker image..."
	docker build -t cors-proxy:latest .
	@echo "✅ Docker image built: cors-proxy:latest"

# Run Docker container
docker-run:
	@echo "🐳 Running Docker container..."
	docker run -p 8080:8080 --rm cors-proxy:latest

# Build and run with Docker
docker: docker-build docker-run

# Docker Compose
docker-compose-up:
	@echo "🐳 Starting with Docker Compose..."
	docker-compose up --build

docker-compose-down:
	docker-compose down

# Clean build artifacts
clean:
	@echo "🧹 Cleaning..."
	rm -rf bin/
	rm -f cors-proxy
	go clean
	@echo "✅ Clean complete"

# Format code
fmt:
	@echo "🎨 Formatting code..."
	go fmt ./...

# Lint code (requires golangci-lint)
lint: fmt
	@echo "🔍 Linting..."
	@command -v golangci-lint > /dev/null 2>&1 || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	golangci-lint run

# Show current configuration
config:
	@echo "⚙️  Current Configuration:"
	@echo "PORT: $${PORT:-8080}"
	@echo "MAX_REQUEST_SIZE: $${MAX_REQUEST_SIZE:-10485760}"
	@echo "REQUEST_TIMEOUT: $${REQUEST_TIMEOUT:-30s}"
	@echo "MAX_REDIRECTS: $${MAX_REDIRECTS:-10}"
	@echo "ALLOWED_ORIGINS: $${ALLOWED_ORIGINS:-*}"
	@echo "ALLOWED_HOSTS: $${ALLOWED_HOSTS:-}"
	@echo "BLOCKED_HOSTS: $${BLOCKED_HOSTS:-}"
	@echo "ALLOW_PRIVATE_NETWORKS: $${ALLOW_PRIVATE_NETWORKS:-false}"
	@echo "RATE_LIMIT_PER_MINUTE: $${RATE_LIMIT_PER_MINUTE:-0}"
	@echo "TRUST_PROXY_HEADERS: $${TRUST_PROXY_HEADERS:-false}"
	@echo "REQUIRE_API_KEY: $${REQUIRE_API_KEY:-false}"
	@echo "DAILY_REQUEST_LIMIT: $${DAILY_REQUEST_LIMIT:-0}"
	@echo "VERBOSE_LOGGING: $${VERBOSE_LOGGING:-false}"
