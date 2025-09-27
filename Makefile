.PHONY: build test clean run deps lint

# Build the application
build:
	go build -o bin/replication-service ./cmd

# Download dependencies
deps:
	go mod tidy
	go mod download

# Run tests
test:
	go test -v ./...

# Run tests with coverage
test-coverage:
	go test -cover ./...

# Run the application
run:
	go run ./cmd

# Clean build artifacts
clean:
	rm -rf bin/

# Lint the code
lint:
	go vet ./...
	go fmt ./...

# Install development tools
install-tools:
	go install github.com/swaggo/swag/cmd/swag@latest

# Generate swagger docs
swagger:
	swag init -g cmd/main.go

# Docker build
docker-build:
	docker build -t table-replication-service .

# Help
help:
	@echo "Available targets:"
	@echo "  build         - Build the application"
	@echo "  deps          - Download dependencies"
	@echo "  test          - Run tests"
	@echo "  test-coverage - Run tests with coverage"
	@echo "  run           - Run the application"
	@echo "  clean         - Clean build artifacts"
	@echo "  lint          - Lint and format code"
	@echo "  swagger       - Generate swagger documentation"
	@echo "  docker-build  - Build Docker image"
