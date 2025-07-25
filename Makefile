APP_NAME=featureflags

.PHONY: build run test clean docker-build docker-run dev dev-build dev-stop scenario-test help

# Build the application
build:
	go build -o $(APP_NAME) ./cmd/main.go

# Run the application locally
run: build
	./$(APP_NAME)

# Run tests
test:
	go test -v ./...

# Run scenario tests
scenario-test:
	docker-compose run --rm scenario-test

# Clean build artifacts
clean:
	rm -f $(APP_NAME) coverage.out coverage.html build-errors.log
	rm -rf tmp/ bin/

# Build Docker image for production
docker-build:
	docker build -t $(APP_NAME) .

# Run with Docker Compose (production)
docker-run:
	docker-compose up --build

# Start development environment with hot reload
dev:
	./scripts/dev.sh

# Build development Docker image
dev-build:
	docker-compose build dev

# Stop development environment
dev-stop:
	docker-compose stop dev db

# Start development environment in background
dev-bg:
	docker-compose up -d dev

# View development logs
dev-logs:
	docker-compose logs -f dev

# Run tests in development environment
dev-test:
	docker-compose run --rm dev go test -v ./...

# Install Air for local development (if not using Docker)
install-air:
	go install github.com/cosmtrek/air@latest

# Run with Air locally (requires PostgreSQL running separately)
air-local:
	air

# Help
help:
	@echo "Available commands:"
	@echo "  build        - Build the application binary"
	@echo "  run          - Run the application locally"
	@echo "  test         - Run tests"
	@echo "  scenario-test - Run comprehensive scenario tests"
	@echo "  clean        - Clean build artifacts"
	@echo ""
	@echo "Docker commands:"
	@echo "  docker-build - Build production Docker image"
	@echo "  docker-run   - Run with Docker Compose (production)"
	@echo ""
	@echo "Development commands:"
	@echo "  dev          - Start development environment with hot reload"
	@echo "  dev-build    - Build development Docker image"
	@echo "  dev-stop     - Stop development environment"
	@echo "  dev-bg       - Start development environment in background"
	@echo "  dev-logs     - View development logs"
	@echo "  dev-test     - Run tests in development environment"
	@echo ""
	@echo "Local development:"
	@echo "  install-air  - Install Air for hot reload"
	@echo "  air-local    - Run with Air locally (requires separate PostgreSQL)" 