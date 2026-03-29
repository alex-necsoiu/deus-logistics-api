.PHONY: help build run test clean docker-up docker-run docker-down migrate sqlc lint install-tools

# Variables
BINARY_NAME=deus-api
GO=go
DOCKER_COMPOSE=docker-compose

help:
	@echo "DEUS Logistics API — Available Commands:"
	@echo ""
	@echo "  make install-tools    Install required tools (sqlc, golang-migrate, mockgen)"
	@echo "  make build            Build the API binary"
	@echo "  make run              Run the API locally (requires db + kafka)"
	@echo "  make test             Run all tests"
	@echo "  make test-coverage    Run tests with coverage report"
	@echo "  make docker-up        Start postgres + kafka + zookeeper with docker-compose"
	@echo "  make docker-run       Alias for docker-up (start full stack)"
	@echo "  make docker-down      Stop all containers"
	@echo "  make migrate-up       Run pending database migrations"
	@echo "  make migrate-down     Rollback last database migration"
	@echo "  make sqlc             Generate sqlc code from queries"
	@echo "  make clean            Clean build artifacts"
	@echo "  make lint             Run golangci-lint (requires installation)"

install-tools:
	@echo "Installing required tools..."
	$(GO) install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	$(GO) install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	$(GO) install github.com/vektra/mockgen/v2/cmd/mockgen@latest

build:
	@echo "Building API binary..."
	$(GO) build -o bin/$(BINARY_NAME) ./cmd/api

run: build
	@echo "Running API..."
	./bin/$(BINARY_NAME)

test:
	@echo "Running tests..."
	$(GO) test -v ./...

test-coverage:
	@echo "Running tests with coverage..."
	$(GO) test -v -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

sqlc:
	@echo "Generating sqlc code..."
	sqlc generate -f internal/postgres/sqlc.yaml

migrate-up:
	@echo "Running pending migrations..."
	migrate -path internal/postgres/migrations -database "postgresql://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=${DB_SSL_MODE}" up

migrate-down:
	@echo "Rolling back last migration..."
	migrate -path internal/postgres/migrations -database "postgresql://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=${DB_SSL_MODE}" down

docker-up:
	@echo "Starting docker-compose services..."
	$(DOCKER_COMPOSE) up -d

docker-run: docker-up
	@echo "Full stack is running at http://localhost:8080"

docker-down:
	@echo "Stopping docker-compose services..."
	$(DOCKER_COMPOSE) down

clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	rm -f coverage.out coverage.html

lint:
	@echo "Running linter..."
	golangci-lint run ./...

.DEFAULT_GOAL := help
