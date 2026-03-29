# 🚢 DEUS Logistics API

Production-grade REST API for managing cargo shipments, vessels, and tracking history with clean architecture and event-driven patterns.

![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat-square&logo=go&logoColor=white)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15-4169E1?style=flat-square&logo=postgresql&logoColor=white)
![Kafka](https://img.shields.io/badge/Kafka-Event--Driven-231F20?style=flat-square&logo=apachekafka&logoColor=white)
![Docker](https://img.shields.io/badge/Docker-Containerized-2496ED?style=flat-square&logo=docker&logoColor=white)
![License](https://img.shields.io/badge/License-MIT-green?style=flat-square)

---

## 📋 Overview

DEUS Logistics API is a production-grade backend service for managing maritime cargo operations. Built with Go 1.21+ and following clean architecture principles with domain-driven design, the API handles cargo shipment lifecycle management, vessel tracking, and immutable transaction history. The system features event-driven architecture with Kafka integration, structured logging, comprehensive error handling, and production-grade operational practices including health checks and testcontainers integration tests.

---

## ✨ Features

| Feature | Description |
|---------|-------------|
| 📦 **Cargo Management** | Create, retrieve, list, and update cargo shipment status with validated state transitions (pending → in_transit → delivered) |
| 🚢 **Vessel Management** | Create, retrieve, list, and update vessel information including location tracking |
| 🔒 **Append-Only Tracking** | Immutable tracking entries per cargo with database-level constraints preventing updates or deletions |
| 📨 **Kafka Event-Driven** | Publish cargo status changes to Kafka topics with in-process consumer goroutine for event persistence |
| 📊 **Structured Logging** | JSON-formatted logs with request tracing via context propagation for observability |
| 🛑 **Centralized Error Mapping** | Consistent error responses with HTTP status codes, error codes, and request IDs |
| ❤️ **Health Checks** | Liveness probe (`/health`) and readiness probe (`/ready`) with dependency validation |
| 🧪 **Integration Testing** | Testcontainers support for real PostgreSQL and Kafka integration tests |

---

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────┐
│                   HTTP Client                        │
└─────────────────────┬───────────────────────────────┘
                      │ REST
┌─────────────────────▼───────────────────────────────┐
│              Transport Layer (Gin)                   │
│         Handlers · DTOs · Error Mapping              │
└─────────────────────┬───────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────┐
│             Application Layer                        │
│        Use Cases · Orchestration · Validation        │
└─────────────────────┬───────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────┐
│                 Domain Layer                         │
│      Entities · State Machine · Business Rules       │
│    cargo/ · vessel/ · tracking/  (per-entity pkg)    │
└─────────────────────┬───────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────┐
│             Infrastructure Layer                     │
│   PostgreSQL (pgx) · Kafka Producer · Consumer       │
└─────────────────────────────────────────────────────┘
```

### Layer Responsibilities

| Layer | Package | Responsibility |
|-------|---------|---|
| **Transport** | `internal/transport/http/` | Gin handlers, DTOs, error mapping, request/response serialization |
| **Application** | `internal/application/cargo/` | Use cases, orchestration, validation, business workflows |
| **Domain** | `internal/domain/cargo/`, `vessel/`, `tracking/` | Entities, state machine, business rules, repository interfaces |
| **Infrastructure** | `internal/postgres/`, `internal/events/` | PostgreSQL (pgx), Kafka producer/consumer, migrations |
| **Service** | `internal/service/` | Vessel and tracking service implementations |

### Event-Driven Flow

```
PATCH /api/v1/cargoes/:id/status
         │
         ▼
  UpdateCargoStatusUseCase.Execute()
         │
         ├─── DB: UPDATE cargoes SET status = ?
         │
         ├─── DB: INSERT INTO tracking_entries (append-only)
         │
         └─── Kafka: publish → "cargo-status-changes" topic
                              │
                              ▼
                   Consumer goroutine (same process)
                              │
                              ▼
                   DB: INSERT INTO cargo_events (append-only)
```

---

## 🛠️ Tech Stack

| Category | Technology | Version/Notes |
|----------|-----------|---|
| **Language** | Go | 1.21+ |
| **HTTP Framework** | Gin Web Framework | Latest |
| **Database** | PostgreSQL | 15-alpine |
| **DB Driver** | pgx with connection pooling | Latest |
| **Migrations** | golang-migrate | Latest |
| **Event Bus** | Apache Kafka | 7.5.0 |
| **Kafka Client** | segmentio/kafka-go | Latest |
| **Logging** | Zerolog | Latest |
| **Testing** | testify + testcontainers | Latest |
| **Containers** | Docker + Docker Compose | Docker Desktop 3.6+ / Engine 23+ |

---

## 📁 Project Structure

```
deus-logistics-api/
├── cmd/
│   └── api/
│       └── main.go                           # Single binary entry point (API + consumer)
├── internal/
│   ├── application/
│   │   └── cargo/
│   │       ├── manager.go                    # Dependency injection wiring
│   │       ├── use_cases.go                  # Create, get, list, update status
│   │       └── interfaces.go                 # Repository interfaces
│   ├── domain/
│   │   ├── cargo/
│   │   │   ├── models.go                     # Cargo, CargoStatus
│   │   │   ├── errors.go                     # Domain-specific errors
│   │   │   ├── events.go                     # StatusChangedEvent
│   │   │   └── repository.go                 # Repository interface
│   │   ├── vessel/
│   │   │   ├── models.go                     # Vessel, VesselStatus
│   │   │   ├── errors.go                     # ErrNotFound, ErrCapacityExceeded
│   │   │   └── repository.go                 # Repository interface
│   │   └── tracking/
│   │       ├── models.go                     # TrackingEntry, AddTrackingInput
│   │       ├── errors.go                     # Validation errors
│   │       └── repository.go                 # Append-only interface
│   ├── service/
│   │   ├── cargo_service.go                  # Cargo service implementation
│   │   ├── vessel_service.go                 # Vessel service implementation
│   │   └── tracking_service.go               # Tracking service implementation
│   ├── postgres/
│   │   ├── db.go                             # Connection pool setup
│   │   ├── migrations/
│   │   │   ├── 000001_init.up.sql            # Create tables
│   │   │   └── 000001_init.down.sql          # Rollback
│   │   ├── queries/
│   │   │   ├── cargo.sql                     # sqlc cargo queries
│   │   │   ├── vessel.sql                    # sqlc vessel queries
│   │   │   └── tracking.sql                  # sqlc tracking queries
│   │   ├── cargo_repo.go                     # sqlc implementation
│   │   ├── vessel_repo.go                    # sqlc implementation
│   │   └── tracking_repo.go                  # sqlc implementation (append-only)
│   ├── events/
│   │   ├── producer.go                       # Kafka producer
│   │   └── consumer.go                       # Kafka consumer (goroutine)
│   ├── transport/
│   │   └── http/
│   │       ├── router.go                     # Route registration
│   │       ├── cargo_handler.go              # Cargo endpoints
│   │       ├── vessel_handler.go             # Vessel endpoints
│   │       ├── tracking_handler.go           # Tracking endpoints
│   │       ├── health_handler.go             # Health check endpoints
│   │       └── error_handler.go              # Error mapping
│   ├── health/
│   │   └── reporter.go                       # Health check logic
│   ├── validation/
│   │   └── validator.go                      # Input validation rules
│   ├── errors/
│   │   └── errors.go                         # Global error types
│   └── config/
│       └── config.go                         # Environment configuration
├── pkg/
│   └── response/
│       ├── constants.go                      # Error codes, response envelopes
│       └── error.go                          # Response wrapper
├── Dockerfile                                # Multi-stage build: golang:1.21-alpine → alpine:3.18
├── docker-compose.yml                        # PostgreSQL, Kafka, Zookeeper, API
├── Makefile                                  # Build targets, tests, migrations
├── .env.example                              # Configuration template
├── .gitignore                                # Git ignore rules
└── README.md                                 # This file
```

---

## 🚀 Getting Started

### Prerequisites

| Requirement | Version | Purpose |
|---|---|---|
| Go | 1.21+ | Build and run locally |
| Docker | 24+ | Run containerized services |
| Docker Compose | 2.x | Service orchestration |
| make | any | Build automation |

### Step 1 — Clone the Repository

```bash
git clone https://github.com/alex-necsoiu/deus-logistics-api.git
cd deus-logistics-api
```

### Step 2 — Configure Environment

```bash
cp .env.example .env
# Edit .env if needed (defaults work with docker-compose)
```

### Step 3 — Start the Full Stack

```bash
make docker-run
```

This starts:
- **PostgreSQL 15** on `localhost:5432`
- **Kafka** on `localhost:9092` (internal: `localhost:29092`)
- **Zookeeper** on `localhost:2181`
- **API** on `localhost:8080`

### Step 4 — Verify Everything is Running

```bash
# Liveness check
curl http://localhost:8080/health

# Readiness check (verifies DB connectivity)
curl http://localhost:8080/ready
```

Expected output:

```json
{
  "status": "healthy",
  "check": {
    "name": "liveness",
    "status": "healthy"
  }
}
```

---

## 📡 API Reference

### Endpoints Summary

| Method | Endpoint | Description |
|---|---|---|
| POST | `/api/v1/cargoes` | Create a cargo shipment |
| GET | `/api/v1/cargoes` | List all cargoes |
| GET | `/api/v1/cargoes/:id` | Get cargo by ID |
| PATCH | `/api/v1/cargoes/:id/status` | Update cargo status |
| POST | `/api/v1/vessels` | Create a vessel |
| GET | `/api/v1/vessels` | List all vessels |
| GET | `/api/v1/vessels/:id` | Get vessel by ID |
| PATCH | `/api/v1/vessels/:id/location` | Update vessel location |
| POST | `/api/v1/cargoes/:id/tracking` | Add tracking entry |
| GET | `/api/v1/cargoes/:id/tracking` | Get tracking history |
| GET | `/health` | Liveness check |
| GET | `/ready` | Readiness check |

### Example 1 — Create Cargo

```bash
curl -X POST http://localhost:8080/api/v1/cargoes \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Container ABC-123",
    "description": "Electronics shipment",
    "weight": 500.0,
    "vessel_id": "550e8400-e29b-41d4-a716-446655440000"
  }'
```

Response (201 Created):

```json
{
  "data": {
    "id": "7c9e6679-7425-40de-944b-e07fc1f90ae7",
    "name": "Container ABC-123",
    "description": "Electronics shipment",
    "weight": 500,
    "status": "pending",
    "vessel_id": "550e8400-e29b-41d4-a716-446655440000",
    "created_at": "2025-01-15T10:30:00Z",
    "updated_at": "2025-01-15T10:30:00Z"
  },
  "meta": {
    "request_id": "a3f1b2c4-e29b-41d4-a716-446655440000"
  }
}
```

### Example 2 — Update Cargo Status

Valid transitions: `pending` → `in_transit` → `delivered`

```bash
curl -X PATCH http://localhost:8080/api/v1/cargoes/7c9e6679-7425-40de-944b-e07fc1f90ae7/status \
  -H "Content-Type: application/json" \
  -d '{
    "status": "in_transit"
  }'
```

Response (200 OK):

```json
{
  "data": {
    "id": "7c9e6679-7425-40de-944b-e07fc1f90ae7",
    "name": "Container ABC-123",
    "status": "in_transit",
    "updated_at": "2025-01-15T10:35:00Z"
  },
  "meta": {
    "request_id": "a3f1b2c4-e29b-41d4-a716-446655440000"
  }
}
```

Invalid transition (422 Unprocessable Entity):

```bash
curl -X PATCH http://localhost:8080/api/v1/cargoes/7c9e6679-7425-40de-944b-e07fc1f90ae7/status \
  -H "Content-Type: application/json" \
  -d '{
    "status": "delivered"
  }'
```

Response (422 Unprocessable Entity):

```json
{
  "error": {
    "code": "INVALID_TRANSITION",
    "message": "Cannot transition from in_transit to delivered only from in_transit",
    "request_id": "a3f1b2c4-e29b-41d4-a716-446655440000"
  }
}
```

### Example 3 — Get Tracking History

```bash
curl http://localhost:8080/api/v1/cargoes/7c9e6679-7425-40de-944b-e07fc1f90ae7/tracking
```

Response (200 OK):

```json
{
  "data": [
    {
      "id": "1a2b3c4d-5e6f-7g8h-9i0j-1k2l3m4n5o6p",
      "cargo_id": "7c9e6679-7425-40de-944b-e07fc1f90ae7",
      "location": "Shanghai Port",
      "status": "pending",
      "note": "Cargo received at origin port",
      "created_at": "2025-01-15T10:30:00Z"
    },
    {
      "id": "2b3c4d5e-6f7g-8h9i-0j1k-2l3m4n5o6p7q",
      "cargo_id": "7c9e6679-7425-40de-944b-e07fc1f90ae7",
      "location": "En route to Rotterdam",
      "status": "in_transit",
      "note": "Vessel departed Shanghai",
      "created_at": "2025-01-15T10:35:00Z"
    }
  ],
  "meta": {
    "request_id": "a3f1b2c4-e29b-41d4-a716-446655440000"
  }
}
```

### Error Response Format

```json
{
  "error": {
    "code": "NOT_FOUND",
    "message": "The requested resource was not found.",
    "request_id": "a3f1b2c4-e29b-41d4-a716-446655440000"
  }
}
```

### Error Codes

| HTTP | Code | Meaning |
|---|---|---|
| 400 | `INVALID_INPUT` | Request validation failed (missing/invalid fields) |
| 404 | `NOT_FOUND` | Resource does not exist |
| 422 | `INVALID_TRANSITION` | State machine rejected the transition |
| 500 | `INTERNAL_ERROR` | Unexpected server error |

---

## 🗄️ Database Schema

### Table Relationships

```
vessels
  └── cargoes (vessel_id → vessels.id  ON DELETE RESTRICT)
        ├── tracking_entries (cargo_id → cargoes.id  APPEND-ONLY 🔒)
        └── cargo_events     (cargo_id → cargoes.id  APPEND-ONLY 🔒)
```

### Key Constraints

- **Cargo Status Constraint:** `CHECK (status IN ('pending', 'in_transit', 'delivered'))`
- **Tracking Append-Only:** Database-level trigger prevents UPDATE/DELETE on `tracking_entries`
- **Events Append-Only:** Database-level trigger prevents UPDATE/DELETE on `cargo_events`
- **Foreign Keys:** Referential integrity with `ON DELETE RESTRICT` for data consistency

### Migrations

| File | Purpose |
|------|---------|
| `000001_init.up.sql` | Create `vessels`, `cargoes`, `tracking_entries`, `cargo_events` tables |
| `000001_init.down.sql` | Drop all tables in reverse dependency order |
| `000002_enforce_append_only_tracking.up.sql` | Add triggers for append-only enforcement |
| `000003_enhance_schema_integrity.up.sql` | Add additional indexes and constraints |

---

## ⚙️ Configuration

### Environment Variables

| Variable | Default | Description |
|---|---|---|
| `DB_HOST` | `postgres` | PostgreSQL hostname |
| `DB_PORT` | `5432` | PostgreSQL port |
| `DB_USER` | `postgres` | Database user |
| `DB_PASSWORD` | `postgres` | Database password |
| `DB_NAME` | `deus_logistics_db` | Database name |
| `DB_SSL_MODE` | `disable` | SSL mode: `disable` or `require` |
| `SERVER_PORT` | `8080` | API listen port |
| `SERVER_ENV` | `development` | Environment: `development` or `production` |
| `KAFKA_BROKER` | `kafka:29092` | Kafka broker address (internal for docker) |
| `KAFKA_TOPIC_STATUS_CHANGES` | `cargo-status-changes` | Kafka topic for status change events |
| `LOG_LEVEL` | `info` | Log level: `debug`, `info`, `warn`, `error` |

---

## 🧪 Testing

### Test Targets

| Command | Description |
|---|---|
| `make test` | Run all tests (unit + integration) |
| `make test-unit` | Unit tests only (no Docker required) |
| `make test-integration` | Integration tests with testcontainers |
| `make test-race` | Run all tests with race detector |
| `make test-coverage` | Generate HTML coverage report (fails if < 80%) |

### Coverage by Package

| Package | Coverage | Status |
|---|---|---|
| `internal/domain/cargo` | ~100% | ✅ Complete |
| `internal/domain/vessel` | ~100% | ✅ Complete |
| `internal/domain/tracking` | ~100% | ✅ Complete |
| `internal/application/cargo` | ~70% | ⚠️ Use case orchestration |
| `internal/service` | ~65% | ⚠️ Service implementations |
| `internal/transport/http` | ~60% | ⚠️ Handler integration |
| `internal/postgres` | ~40% | ⚠️ Testcontainers tests |

### Integration Tests

Integration tests use **testcontainers-go** to spin up real PostgreSQL and Kafka instances:

```bash
# Requires Docker to be running
go test ./internal/postgres/... -v -run TestCargoRepository

# With race detector
make test-race
```

---

## 🔧 Development

### Makefile Targets

| Target | Description |
|---|---|
| `make help` | Display all available targets |
| `make install-tools` | Install sqlc, golang-migrate, mockgen |
| `make build` | Compile API binary (with version embedding) |
| `make run` | Build and run locally |
| `make generate` | Run go generate (mockgen) |
| `make fmt` | Format all Go files |
| `make fmt-check` | Check formatting (CI-safe, exits 1 if issues) |
| `make vet` | Run static analysis (go vet) |
| `make lint` | Run golangci-lint |
| `make sqlc` | Generate type-safe code from SQL queries |
| `make test` | Run all tests |
| `make test-unit` | Unit tests only |
| `make test-race` | Tests with race detector |
| `make test-coverage` | Coverage report (enforces 80% minimum) |
| `make docker-build` | Build Docker image |
| `make docker-run` / `make dev-up` | Start full Docker stack |
| `make docker-down` / `make dev-down` | Stop all containers |
| `make docker-logs` / `make logs` | Tail container logs |
| `make docker-ps` / `make ps` | Show container status |
| `make migrate-up` | Run pending migrations |
| `make migrate-down` | Rollback last migration |
| `make migrate-create` | Create new migration pair |
| `make clean` | Remove build artifacts |

### Structured Logging Example

All logs are emitted as JSON for machine parsing:

```json
{
  "level": "info",
  "time": "2025-01-15T10:30:45Z",
  "caller": "application/cargo/update_status.go:95",
  "cargo_id": "7c9e6679-7425-40de-944b-e07fc1f90ae7",
  "old_status": "pending",
  "new_status": "in_transit",
  "message": "business_event:cargo_status_updated"
}
```

### Code Generation

Generate mocks and sqlc code:

```bash
make generate    # Runs go generate ./...
make sqlc        # Regenerate sqlc from SQL queries
```

---

## 🐳 Docker

### Services

| Service | Image | Port | Purpose |
|---|---|---|---|
| `api` | Built from `Dockerfile` | `8080` | DEUS Logistics API |
| `postgres` | `postgres:15-alpine` | `5432` | Primary database |
| `kafka` | `confluentinc/cp-kafka:7.5.0` | `9092`, `29092` | Event streaming bus |
| `zookeeper` | `confluentinc/cp-zookeeper:7.5.0` | `2181` | Kafka coordination |

### Multi-Stage Dockerfile

```dockerfile
# Stage 1: Build (golang:1.21-alpine)
# - Install dependencies
# - Compile with ldflags (version, build time, commit)
# - Result: ~18MB binary

# Stage 2: Runtime (alpine:3.18)
# - Copy binary only
# - Health checks
# - Final image: ~20MB
```

### Health Checks

All services include health checks:

```yaml
healthcheck:
  test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
  interval: 10s
  timeout: 5s
  retries: 3
```

---

## 📄 License

MIT License — Copyright (c) 2025 Alex Necsoiu

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software.

---

## 👤 Author

**Alex Necsoiu**

- GitHub: [@alex-necsoiu](https://github.com/alex-necsoiu)
- LinkedIn: [alex-necsoiu](https://www.linkedin.com/in/alex-necsoiu/)
- Email: alex@example.com

---

> 💡 **For questions or contributions, please open an issue or submit a pull request on GitHub.**
