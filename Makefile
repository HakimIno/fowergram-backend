.PHONY: dev build test clean

# Development with hot reload using Air
dev:
	@echo "Starting development server with hot reload..."
	@command -v air > /dev/null 2>&1 || (echo "Installing air..." && go install github.com/cosmtrek/air@latest)
	@air

# Build the application
build:
	@echo "Building application..."
	@go build -o ./tmp/main ./cmd/api

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./...

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf ./tmp/main
	@rm -rf ./tmp/air.log

.DEFAULT_GOAL := dev 