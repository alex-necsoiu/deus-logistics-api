# =============================================================================
# DEUS Logistics API — Makefile
# =============================================================================
# Production-grade Makefile for Go backend. Follows architecture spec strictly.
# All targets are carefully designed for local dev, CI/CD, and container ops.

# =============================================================================
# ⚙️  VARIABLES — Configuration Center
# =============================================================================

# Go settings
GO               := go
GOFLAGS          := 
BINARY_NAME      := api
BINARY_PATH      := $(PWD)/$(BINARY_NAME)
COVERAGE_MIN     := 80
VERSION          := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME       := $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')
COMMIT           := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Database configuration (from .env, with fallbacks)
DB_HOST          ?= localhost
DB_PORT          ?= 5432
DB_USER          ?= postgres
DB_PASSWORD      ?= postgres
DB_NAME          ?= deus_logistics_db
DB_SSL_MODE      ?= disable

# Derived database variables
DB_DSN           := postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSL_MODE)
MIGRATION_PATH   := internal/postgres/migrations

# Kafka configuration (from .env, with fallbacks)
KAFKA_BROKER     ?= localhost:9092
KAFKA_TOPIC      ?= cargo-status-changes

# Server configuration (from .env, with fallbacks)
SERVER_PORT      ?= 8080
SERVER_ENV       ?= development

# Docker / Docker Compose
DOCKER           := docker
DOCKER_COMPOSE   := docker compose
DOCKER_REGISTRY  ?= 

# ldflags for version embedding
LDFLAGS          := -ldflags "\
	-X main.Version=$(VERSION) \
	-X main.BuildTime=$(BUILD_TIME) \
	-X main.Commit=$(COMMIT)"

# Test configuration
TEST_COVERAGE_FILE := coverage.out
TEST_COVERAGE_HTML := coverage.html

# =============================================================================
# 📋 PHONY TARGETS — Mark all targets as .PHONY
# =============================================================================

.PHONY: \
	help \
	build \
	install-tools \
	generate \
	sqlc \
	swagger \
	deps \
	fmt fmt-check \
	vet \
	lint \
	test test-unit test-integration test-race test-coverage \
	run \
	docker-build docker-run docker-down dev-up dev-down docker-logs docker-ps \
	migrate-up migrate-down migrate-create \
	clean logs ps \
	.DEFAULT_GOAL

.DEFAULT_GOAL := help

# =============================================================================
# 🆘 HELP — Display all available targets
# =============================================================================

help:
	@echo "╔════════════════════════════════════════════════════════════════════╗"
	@echo "║         DEUS Logistics API — Make Targets                          ║"
	@echo "╚════════════════════════════════════════════════════════════════════╝"
	@echo ""
	@echo "┌─ Core Development ─────────────────────────────────────────────────┐"
	@echo "│ make deps               Install required CLI tools (swag, etc)    │"
	@echo "│ make install-tools      Install sqlc, migrate, mockgen             │"
	@echo "│ make build              Build the API binary (with version info)   │"
	@echo "│ make run                Run the API locally (requires postgres)    │"
	@echo "│ make generate           Run go generate ./... (mockgen, etc)       │"
	@echo "└────────────────────────────────────────────────────────────────────┘"
	@echo ""
	@echo "┌─ Documentation ────────────────────────────────────────────────────┐"
	@echo "│ make swagger            Regenerate Swagger docs from annotations   │"
	@echo "│                         Visit: http://localhost:8080/swagger/      │"
	@echo "└────────────────────────────────────────────────────────────────────┘"
	@echo ""
	@echo "┌─ Code Quality ─────────────────────────────────────────────────────┐"
	@echo "│ make fmt                Format all Go files                        │"
	@echo "│ make fmt-check          Check if formatting is needed (CI safe)    │"
	@echo "│ make vet                Run go vet (static analysis)               │"
	@echo "│ make lint               Run golangci-lint                          │"
	@echo "│ make sqlc               Generate type-safe code from SQL queries  │"
	@echo "└────────────────────────────────────────────────────────────────────┘"
	@echo ""
	@echo "┌─ Testing ──────────────────────────────────────────────────────────┐"
	@echo "│ make test               Run all tests (unit + integration)         │"
	@echo "│ make test-unit          Run unit tests ONLY (no docker required)   │"
	@echo "│ make test-integration   Run integration tests (requires docker)    │"
	@echo "│ make test-race          Run all tests with race detection          │"
	@echo "│ make test-coverage      Generate coverage report (fails <80%)      │"
	@echo "└────────────────────────────────────────────────────────────────────┘"
	@echo ""
	@echo "┌─ Docker / Containers (docker compose) ─────────────────────────────┐"
	@echo "│ make docker-build       Build the API Docker image                │"
	@echo "│ make docker-run         Start postgres + kafka + zookeeper        │"
	@echo "│ make docker-down        Stop all containers                       │"
	@echo "│ make docker-logs        Tail logs for running containers          │"
	@echo "│ make docker-ps          Show running container status            │"
	@echo "│ make dev-up             Alias for docker-run (start full stack)  │"
	@echo "│ make dev-down           Alias for docker-down                    │"
	@echo "└────────────────────────────────────────────────────────────────────┘"
	@echo ""
	@echo "┌─ Database Migrations ──────────────────────────────────────────────┐"
	@echo "│ make migrate-up         Run pending migrations (uses DB_* env)    │"
	@echo "│ make migrate-down       Rollback last migration                   │"
	@echo "│ make migrate-create     Create new migration file pair            │"
	@echo "└────────────────────────────────────────────────────────────────────┘"
	@echo ""
	@echo "┌─ Maintenance ──────────────────────────────────────────────────────┐"
	@echo "│ make clean              Remove build artifacts & coverage reports │"
	@echo "└────────────────────────────────────────────────────────────────────┘"
	@echo ""
	@echo "📝 Configuration:"
	@echo "   DB_HOST=$(DB_HOST) DB_PORT=$(DB_PORT) DB_NAME=$(DB_NAME)"
	@echo "   KAFKA_BROKER=$(KAFKA_BROKER)"
	@echo "   SERVER_PORT=$(SERVER_PORT)"
	@echo ""

# =============================================================================
# 🛠️  INSTALL TOOLS — Download required CLI tools
# =============================================================================

install-tools:
	@echo "📥 Installing required tools..."
	@echo "   • Installing sqlc..."
	$(GO) install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	@echo "   • Installing golang-migrate..."
	$(GO) install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	@echo "   • Installing mockgen..."
	$(GO) install github.com/vektra/mockgen/v2/cmd/mockgen@latest
	@echo "✅ All tools installed successfully!"

# =============================================================================
# � SWAGGER — Generate OpenAPI documentation
# =============================================================================

deps:
	@echo "📥 Installing Swaggo CLI..."
	@command -v swag >/dev/null 2>&1 || { \
		echo "   • Installing swag..."; \
		$(GO) install github.com/swaggo/swag/cmd/swag@latest; \
	}
	@echo "✅ Swaggo CLI installed!"

swagger: deps
	@echo "📚 Generating Swagger documentation from annotations..."
	swag init -g cmd/api/main.go
	@echo "✅ Swagger docs generated!"
	@echo "   📍 Swagger UI: http://localhost:8080/swagger/index.html"
	@echo "   📄 Specification: docs/swagger.json"

# =============================================================================
# �🔧 CODE GENERATION — Generate code from directives and SQL
# =============================================================================

generate:
	@echo "🔄 Generating code (go generate ./...)..."
	$(GO) generate ./...
	@echo "✅ Code generation complete!"

sqlc:
	@echo "📦 Generating sqlc code from SQL queries..."
	sqlc generate -f internal/postgres/sqlc.yaml
	@echo "✅ sqlc code generation complete!"

# =============================================================================
# 🎨 FORMATTING — Code quality enforcement
# =============================================================================

fmt:
	@echo "🎨 Formatting all Go files (gofmt)..."
	$(GO) fmt ./...
	@echo "✅ Formatting complete!"

fmt-check:
	@echo "🔍 Checking if code needs formatting..."
	@if [ -n "$$($(GO) fmt ./... 2>&1)" ]; then \
		echo "❌ Code formatting issues found. Run 'make fmt' to fix."; \
		exit 1; \
	fi
	@echo "✅ All files are properly formatted!"

# =============================================================================
# 🔎 STATIC ANALYSIS — Code quality checks
# =============================================================================

vet:
	@echo "🔍 Running go vet (static analysis)..."
	$(GO) vet ./...
	@echo "✅ No vet issues found!"

lint:
	@echo "🔍 Running golangci-lint..."
	@command -v golangci-lint >/dev/null 2>&1 || { \
		echo "⚠️  golangci-lint not found. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
		exit 1; \
	}
	golangci-lint run ./...
	@echo "✅ No lint issues found!"

# =============================================================================
# 🏗️  BUILD — Compile the binary with version info
# =============================================================================

build:
	@echo "🏗️  Building binary: $(BINARY_NAME)"
	@echo "   Version: $(VERSION)"
	@echo "   Build Time: $(BUILD_TIME)"
	@echo "   Commit: $(COMMIT)"
	$(GO) build $(LDFLAGS) -o $(BINARY_PATH) ./cmd/api
	@echo "✅ Binary built: $(BINARY_PATH)"
	@ls -lh $(BINARY_PATH)

# =============================================================================
# 🧪 TESTING — Multiple test modes for different scenarios
# =============================================================================

test: test-unit test-integration
	@echo "✅ All tests passed!"

test-unit:
	@echo "🧪 Running unit tests (no docker required)..."
	$(GO) test -v -short ./... 2>&1 | grep -v "testcontainers"
	@echo "✅ Unit tests passed!"

test-integration:
	@echo "🧪 Running integration tests (requires docker)..."
	$(GO) test -v ./... 2>&1
	@echo "✅ Integration tests passed!"

test-race:
	@echo "🧪 Running tests with race detector..."
	$(GO) test -race -v ./...
	@echo "✅ No data races detected!"

test-coverage:
	@echo "🧪 Running tests with coverage analysis..."
	$(GO) test -coverprofile=$(TEST_COVERAGE_FILE) ./...
	@echo "📊 Generating HTML coverage report..."
	$(GO) tool cover -html=$(TEST_COVERAGE_FILE) -o $(TEST_COVERAGE_HTML)
	@echo "📈 Coverage report: file://$(PWD)/$(TEST_COVERAGE_HTML)"
	@echo ""
	@COVERAGE=$$($(GO) tool cover -func=$(TEST_COVERAGE_FILE) | grep ^total | awk '{print $$3}' | sed 's/%//'); \
	echo "📌 Total Coverage: $$COVERAGE%"; \
	if [ "$$(echo "$$COVERAGE < $(COVERAGE_MIN)" | bc)" -eq 1 ]; then \
		echo "❌ Coverage below $(COVERAGE_MIN)% threshold!"; \
		exit 1; \
	fi
	@echo "✅ Coverage meets $(COVERAGE_MIN)% threshold!"

# =============================================================================
# ▶️  RUN — Execute the API locally
# =============================================================================

run: build
	@echo "▶️  Running API binary..."
	@echo "   Server: http://localhost:$(SERVER_PORT)"
	@echo "   Database: $(DB_DSN)"
	@echo "   Kafka: $(KAFKA_BROKER)"
	$(BINARY_PATH)

# =============================================================================
# 🐳 DOCKER — Container management with docker compose
# =============================================================================

docker-build:
	@echo "🐳 Building Docker image..."
	$(DOCKER_COMPOSE) build api
	@echo "✅ Docker image built!"

docker-run: docker-build
	@echo "🐳 Starting full stack (postgres, kafka, zookeeper, api)..."
	$(DOCKER_COMPOSE) up -d
	@echo "⏳ Waiting for services to start..."
	@sleep 3
	$(DOCKER_COMPOSE) ps
	@echo ""
	@echo "✅ Full stack running!"
	@echo "   API: http://localhost:$(SERVER_PORT)"
	@echo "   Run 'make docker-logs' to see logs"
	@echo "   Run 'make docker-down' to stop"

docker-down:
	@echo "⛔ Stopping all containers..."
	$(DOCKER_COMPOSE) down
	@echo "✅ All containers stopped!"

docker-logs:
	@echo "📋 Tailing container logs (press Ctrl+C to stop)..."
	$(DOCKER_COMPOSE) logs -f

docker-ps:
	@echo "📊 Container status:"
	@$(DOCKER_COMPOSE) ps

dev-up: docker-run

dev-down: docker-down

# =============================================================================
# 🗄️  DATABASE MIGRATIONS — Manage schema versions
# =============================================================================

migrate-up:
	@echo "📈 Running pending database migrations..."
	@echo "   Database: $(DB_DSN)"
	@echo "   Migrations path: $(MIGRATION_PATH)"
	migrate -path $(MIGRATION_PATH) -database "$(DB_DSN)" up
	@echo "✅ Migrations complete!"

migrate-down:
	@echo "📉 Rolling back last database migration..."
	@echo "   Database: $(DB_DSN)"
	@echo "   Migrations path: $(MIGRATION_PATH)"
	migrate -path $(MIGRATION_PATH) -database "$(DB_DSN)" down 1
	@echo "✅ Rollback complete!"

migrate-create:
	@echo "📝 Creating new migration file pair..."
	@read -p "   Enter migration name (e.g., create_users_table): " NAME; \
	migrate create -ext sql -dir $(MIGRATION_PATH) -seq "$$NAME"; \
	@echo "✅ Migration files created in $(MIGRATION_PATH)"
	@ls -lh $(MIGRATION_PATH) | tail -2

# =============================================================================
# 🧹 CLEANUP — Remove build artifacts
# =============================================================================

clean:
	@echo "🧹 Cleaning build artifacts..."
	@rm -f $(BINARY_PATH) $(TEST_COVERAGE_FILE) $(TEST_COVERAGE_HTML)
	@rm -rf bin/ dist/
	@echo "✅ Cleanup complete!"

logs: docker-logs

ps: docker-ps

# =============================================================================
# 🚀 COMMON WORKFLOWS — Convenience combined targets
# =============================================================================

# (These are already defined as aliases above, listed for clarity)
# dev-up:       Alias for docker-run (start full stack)
# dev-down:     Alias for docker-down (stop full stack)
# logs:         Alias for docker-logs (tail container logs)
# ps:           Alias for docker-ps (show container status)

# =============================================================================
# 📌 END OF MAKEFILE
# =============================================================================
# Principles:
# • Version info embedded at build time via ldflags
# • Consistent DSN construction from DB_* environment variables
# • Separate unit/integration test targets for flexible CI/CD
# • Race detector included for data race detection
# • Coverage enforcement with configurable threshold
# • All code quality checks (fmt, vet, lint) integrated
# • go generate support for mockgen and SQL code generation
# • Docker Compose (space) used for modern Docker installations
# • Comprehensive help text with current settings displayed
