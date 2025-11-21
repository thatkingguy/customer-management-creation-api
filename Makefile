.PHONY: build run test test-coverage migrate-up migrate-down docker-build docker-run lint fmt clean

# Build the application
build:
	go build -o bin/customer-service ./cmd/server

# Run the application locally
run:
	go run ./cmd/server

# Run with hot reload (requires air: go install github.com/air-verse/air@latest)
dev:
	@if command -v air > /dev/null; then \
		air; \
	else \
		echo "Air not installed. Installing..."; \
		go install github.com/air-verse/air@latest; \
		echo "Air installed! Run 'make dev' again."; \
	fi

# Run tests
test:
	go test -v ./...

# Generate test coverage
test-coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Run database migrations up
migrate-up:
	@echo "Running migrations..."
	@bash scripts/run-migrations.sh

# Run specific migration file
migrate-file:
	@if [ -z "$(FILE)"; then \
		echo "Usage: make migrate-file FILE=path/to/migration.sql"; \
	else \
		bash scripts/run-migrations.sh $(FILE); \
	fi

# Run database migrations down
migrate-down:
	@echo "Rolling back migrations..."
	@if [ -f scripts/migrate.sh ]; then \
		bash scripts/migrate.sh down; \
	else \
		echo "Migration script not found. Please run migrations manually."; \
	fi

# Build Docker image
docker-build:
	docker build -f deployments/Dockerfile -t customer-service:latest .

# Run Docker container
docker-run:
	docker-compose -f deployments/docker-compose.yml up -d

# Stop Docker container
docker-stop:
	docker-compose -f deployments/docker-compose.yml down

# Run linter
lint:
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed. Install it with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

# Format code
fmt:
	go fmt ./...
	goimports -w .

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f coverage.out coverage.html

# Install dependencies
deps:
	go mod download
	go mod tidy

# Verify dependencies
verify:
	go mod verify

