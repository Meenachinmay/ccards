# Add these to your Makefile
.PHONY: test test-unit test-integration test-verbose test-coverage

# Run all tests
test:
	@echo "Running all tests..."
	@go test -v ./...

# Run unit tests only
test-unit:
	@echo "Running unit tests..."
	@go test -v -short ./...

# Run integration tests only
test-integration:
	@echo "Running integration tests..."
	@go test -v -run Integration ./...

# Run tests with verbose output
test-verbose:
	@echo "Running tests with verbose output..."
	@go test -v -race ./...

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test -v -race -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html

# Setup test environment
test-setup:
	@echo "Setting up test environment..."
	@docker compose -f docker-local.compose.yml up -d postgres_test redis
	@echo "Waiting for services to be ready..."
	@sleep 5

# Teardown test environment
test-teardown:
	@echo "Tearing down test environment..."
	@docker compose -f docker-local.compose.yml down postgres_test redis

# Run tests in Docker
test-docker: test-setup
	@echo "Running tests in Docker environment..."
	@APP_ENVIRONMENT=test go test -v ./...
	@make test-teardown

# Run
run:
	go run ./cmd/server/main.go

docker-stop:
	@echo "Tearing down test environment..."
	@docker compose -f docker-local.compose.yml down
