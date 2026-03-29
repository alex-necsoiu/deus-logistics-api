# DEUS Logistics API

A Go-based REST API for managing cargo shipments, vessels, and tracking history. Implements clean architecture with domain-driven design, event-driven patterns, and production-grade operational practices.

---

## Features

- **Cargo Management** — Create, retrieve, list, and update cargo shipment status with validated state transitions
- **Vessel Management** — Create, retrieve, list, and update vessel location
- **Append-Only Tracking History** — Immutable tracking entries per cargo shipment with database-level constraints
- **Event-Driven Architecture** — Kafka producer/consumer for cargo status change events
- **Structured Logging** — JSON-formatted logs with request tracing via context propagation
- **REST API** — HTTP handlers with input validation and consistent error responses
- **Docker Support** — Full containerization with PostgreSQL, Kafka, and Zookeeper

---

## Architecture

This project follows **Clean Architecture** with clear separation of concerns across layers:

```
HTTP Request
    ↓
HTTP Handler (Transport)
    ↓
Use Case (Application)
    ↓
Repository (Infrastructure)
    ↓
Domain Model (Business Logic)
```

### Layers

| Layer | Responsibility | Location |
|-------|---|---|
| **Transport** | HTTP handlers, DTOs, error mapping | `internal/transport/http/` |
| **Application** | Use cases, orchestration, business workflows | `internal/application/cargo/` |
| **Domain** | Entities, value objects, business rules, repository interfaces | `internal/domain/` |
| **Infrastructure** | PostgreSQL repositories, Kafka producer/consumer | `internal/postgres/`, `internal/events/` |

### Key Architectural Decisions

- **Domain-per-entity structure:** Cargo, vessel, and tracking each have isolated domain packages
- **Repository pattern:** Data access abstraction enables testing and storage flexibility
- **Event sourcing:** Kafka publishes cargo status changes for downstream systems
- **Append-only tracking:** Database constraints enforce immutability of tracking history
- **Context propagation:** Request IDs flow through all layers for distributed tracing

---

## Tech Stack

| Component | Technology | Version |
|-----------|-----------|---------|
| **Language** | Go | 1.21+ |
| **HTTP Framework** | Gin Web Framework | v1.9.0 |
| **Database** | PostgreSQL | 15-alpine |
| **Database Driver** | pgx with connection pooling | v5.3.0 |
| **Event Bus** | Apache Kafka | 7.5.0 |
| **Kafka Client** | segmentio/kafka-go | v0.4.38 |
| **Logging** | Zerolog | v1.29.0 |
| **Testing** | testify, testcontainers | latest |
| **Containers** | Docker & Docker Compose | latest |

---

## Prerequisites

- **Docker & Docker Compose** — Required for running PostgreSQL, Kafka, and Zookeeper
- **Go 1.21+** — For local development and testing
- **.env configuration** — Environment variables for database and Kafka connection

---

## Getting Started

### 1. Clone and Setup

```bash
git clone https://github.com/alex-necsoiu/deus-logistics-api
cd deus-logistics-api

# Create environment file
cp .env.example .env
```

### 2. Start Services

```bash
docker-compose up --build
```

This starts:
- **PostgreSQL 15** on `localhost:5432`
- **Kafka** on `localhost:9092`
- **Zookeeper** on `localhost:2181`
- **API** on `localhost:8080`

### 3. Verify Health

```bash
# Health check (database connectivity)
curl http://localhost:8080/health

# Readiness check (ready to serve requests)
curl http://localhost:8080/ready
```

---

## API Endpoints

### Health & Status

```
GET  /health              # Service health status
GET  /ready               # Readiness probe (DB connectivity)
```

### Cargo Management

```
POST   /api/v1/cargoes                    # Create cargo
GET    /api/v1/cargoes                    # List all cargo
GET    /api/v1/cargoes/:id                # Get cargo by ID
PATCH  /api/v1/cargoes/:id/status         # Update cargo status
```

### Vessel Management

```
POST   /api/v1/vessels                    # Create vessel
GET    /api/v1/vessels                    # List all vessels
GET    /api/v1/vessels/:id                # Get vessel by ID
PATCH  /api/v1/vessels/:id/location       # Update vessel location
```

### Tracking History

```
POST   /api/v1/cargoes/:id/tracking       # Add tracking entry
GET    /api/v1/cargoes/:id/tracking       # Get tracking history
```

---

## Domain Models

### Cargo

Represents a shipment of goods assigned to a vessel.

**Status Lifecycle:**
- `pending` — Initial state, not yet in transit
- `in_transit` — Currently being transported
- `delivered` — Reached destination

**Valid Transitions:**
- `pending` → `in_transit`
- `in_transit` → `delivered`

Invalid transitions are rejected at the domain layer with validation errors.

### Vessel

Represents a transport vessel with capacity and location tracking.

### Tracking Entry

Immutable audit trail of cargo location and status changes. Supports append-only operations with database-level constraints to prevent modification or deletion.

---

## Development

### Running Tests

```bash
# Run full test suite with coverage
go test ./... -v -cover

# Run specific package tests
go test ./internal/domain/cargo/ -v

# Generate coverage report
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Build

```bash
go build -o api ./cmd/api

# or with Docker
docker build -t deus-api .
```

### Logging

All services use **Zerolog** for structured JSON logging. Logs include:
- Timestamp and caller location
- Request ID for tracing
- Structured fields (cargo_id, vessel_id, status, etc.)
- Error context when failures occur

Example log output:
```json
{
  "level": "info",
  "time": "2026-03-29T15:23:45Z",
  "caller": "internal/postgres/cargo_repo.go:69",
  "cargo_id": "550e8400-e29b-41d4-a716-446655440000",
  "message": "cargo inserted into database"
}
```

---

## Project Structure

```
.
├── api                              # Built binary
├── cmd/
│   └── api/
│       └── main.go                  # Application entry point
├── internal/
│   ├── application/cargo/           # Use cases for cargo management
│   │   ├── create_cargo.go
│   │   ├── get_cargo.go
│   │   ├── list_cargos.go
│   │   ├── update_status.go
│   │   └── interfaces.go
│   ├── config/                      # Configuration management
│   ├── domain/
│   │   ├── cargo/                   # Cargo bounded context
│   │   │   ├── models.go
│   │   │   ├── errors.go
│   │   │   ├── events.go
│   │   │   └── repository.go        # Interface definition
│   │   ├── vessel/                  # Vessel bounded context
│   │   └── tracking/                # Tracking bounded context
│   ├── events/
│   │   ├── producer.go              # Kafka event publisher
│   │   └── consumer.go              # Kafka event consumer
│   ├── postgres/                    # PostgreSQL implementations
│   │   ├── cargo_repo.go
│   │   ├── vessel_repo.go
│   │   ├── tracking_repo.go
│   │   ├── event_repo.go
│   │   ├── migrations/              # SQL migrations
│   │   └── db.go
│   ├── service/                     # Domain services (vessel, tracking)
│   ├── transport/http/              # HTTP transport layer
│   │   ├── cargo_handler.go
│   │   ├── vessel_handler.go
│   │   ├── tracking_handler.go
│   │   ├── router.go
│   │   └── dto.go
│   ├── health/                      # Health check reporter
│   ├── config/                      # Environment configuration
│   └── validation/                  # Input validators
├── pkg/
│   ├── response/                    # JSON response helpers
│   └── validator/                   # Validation utilities
├── docs/
│   └── TEST_RESULTS.md              # Test coverage report
├── Dockerfile
├── docker-compose.yml
├── go.mod
├── go.sum
└── README.md
```

---

## Error Handling

The API returns consistent error responses with HTTP status codes and error details:

```json
{
  "error": {
    "code": "INVALID_INPUT",
    "message": "cargo_id is required and must be a valid UUID",
    "request_id": "550e8400-e29b-41d4-a716-446655440000"
  }
}
```

**Common Error Codes:**
- `INVALID_INPUT` — Validation failure (HTTP 400)
- `NOT_FOUND` — Resource does not exist (HTTP 404)
- `CONFLICT` — Invalid state transition (HTTP 409)
- `INTERNAL_ERROR` — Unexpected server error (HTTP 500)

---

## Configuration

Environment variables via `.env`:

```env
# Database
DB_HOST=postgres
DB_PORT=5432
DB_USER=postgres
DB_PASS=password
DB_NAME=deus_logistics
DB_SSLMODE=disable

# Server
SERVER_PORT=8080
SERVER_ENV=development

# Kafka
KAFKA_BROKERS=kafka:9092
KAFKA_TOPIC_EVENTS=cargo.events
```

---

## Testing

- **80+ unit tests** across domain, application, and service layers
- **Integration tests** using testcontainers for PostgreSQL
- **Test coverage:** 35%+ overall, with core packages at 100%
  - Domain models: 100%
  - Errors: 100%
  - Validation: 100%
  - Application use cases: 69.5%
  - Services: 63.0%

---

## Performance Characteristics

- **Database Pooling:** pgxpool with connection reuse
- **Request Tracing:** Context-based request ID propagation
- **Structured Logging:** Zero-allocation JSON marshaling (Zerolog)
- **Event Publishing:** Fire-and-forget Kafka pattern for non-blocking operations

---

## Production Readiness

- ✅ Clean Architecture with clear separation of concerns
- ✅ Comprehensive error handling and validation
- ✅ Structured logging for observability
- ✅ Request ID tracing for distributed debugging
- ✅ Database health checks and migrations
- ✅ Docker containerization for consistency
- ✅ 80+ tests with high coverage in critical paths
- ✅ Graceful shutdown with context cancellation

---

## Contributing

1. Follow the architecture layers — don't cross boundaries
2. Write tests for new features (domain logic, use cases)
3. Use Zerolog for logging, never `fmt.Println`
4. Keep handler logic thin — push business logic to domain
5. Document domain constraints in model comments

---

## License

MIT License

Copyright (c) 2025 Alex Necsoiu

---

## Authors

- **Alex Necsoiu** — Implementation