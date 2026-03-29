# ✅ DEUS Logistics API — Backend Architecture Specification (Senior Architecture)

> **Status:** Approved
> **Scope:** Full backend implementation — Logistics API Challenge
> **Audience:** Backend engineers + AI coding assistants
> **Rule:** This is the **single source of truth** for all development. ALL code must follow this spec exactly.

---

## 1. Overview

The DEUS Logistics API is a production-grade backend service for a logistics company that ships goods via vessels.

**Core entities:**

- **Vessels** — ships that carry cargo (name, capacity, current_location)
- **Cargoes** — goods being transported (status: pending → in_transit → delivered)
- **Tracking** — immutable movement history per cargo (append-only, never overwritten)
- **Events** — Kafka messages on cargo status change → stored in `cargo_events` table

**Architecture approach:**

- Clean Architecture with domain-per-entity packages (Pandora Exchange pattern)
- Single binary (API + in-process Kafka consumer goroutine)
- sqlc for type-safe, compile-time-verified database access
- Table-driven tests with mockgen mocks + testcontainers for integration tests

---

## 2. Architectural Goals

| Goal | Description |
|---|---|
| **Clean Architecture** | Domain knows nothing about infrastructure |
| **Domain per Package** | Each entity owns its models, errors, and interfaces in one package |
| **Auditability** | Immutable tracking logs — append-only, never overwritten |
| **Testability** | TDD, 80%+ coverage, mockgen, testcontainers |
| **Event-Driven (Bonus)** | Kafka producer on status change + in-process consumer goroutine |
| **Reproducibility** | Docker + docker-compose, one-command startup |

---

## 3. Technology Stack (Non-Negotiable)

| Category | Technology |
|---|---|
| Language | Go 1.21+ |
| HTTP Framework | Gin |
| Database | PostgreSQL 15 |
| Data Access | **sqlc** (SQL-first — no GORM, no raw `db.Query`) |
| DB Migrations | golang-migrate |
| Event Bus | Kafka (`segmentio/kafka-go`) |
| Logging | Zerolog |
| IDs | `github.com/google/uuid` |
| Testing | Go testing + testify + mockgen + testcontainers |
| Containers | Docker + docker-compose |
| Config | ENV vars (`.env` + `.env.example`) |

---

## 4. Folder Structure (Pandora Pattern)

```
/deus-logistics-api
├── cmd/
│   └── api/
│       └── main.go                    # Single binary — API + consumer goroutine
├── internal/
│   ├── domain/
│   │   ├── cargo/                     # Cargo bounded context
│   │   │   ├── models.go              # Cargo, CargoStatus, CreateCargoInput
│   │   │   ├── errors.go              # ErrNotFound, ErrInvalidStatus, ErrInvalidInput
│   │   │   ├── events.go              # CargoEvent, StatusChangedEvent, EventPublisher/EventRepository interfaces
│   │   │   ├── repository.go          # Repository interface
│   │   │   └── service.go             # Service interface
│   │   ├── vessel/                    # Vessel bounded context
│   │   │   ├── models.go              # Vessel, CreateVesselInput
│   │   │   ├── errors.go              # ErrNotFound, ErrCapacityExceeded
│   │   │   ├── repository.go          # Repository interface
│   │   │   └── service.go             # Service interface
│   │   └── tracking/                  # Tracking bounded context
│   │       ├── models.go              # TrackingEntry, AddTrackingInput
│   │       ├── errors.go              # ErrInvalidEntry
│   │       ├── repository.go          # Repository interface
│   │       └── service.go             # Service interface
│   ├── service/                       # Service implementations
│   │   ├── cargo_service.go           # Implements cargo.Service
│   │   ├── cargo_service_test.go      # Unit tests (mockgen mocks)
│   │   ├── vessel_service.go          # Implements vessel.Service
│   │   ├── vessel_service_test.go
│   │   ├── tracking_service.go        # Implements tracking.Service
│   │   └── tracking_service_test.go
│   ├── postgres/                      # Infrastructure: sqlc implementations
│   │   ├── migrations/
│   │   │   ├── 000001_init.up.sql     # Create all tables + indexes
│   │   │   └── 000001_init.down.sql   # Drop all tables in reverse order
│   │   ├── queries/
│   │   │   ├── cargo.sql              # sqlc queries for cargo
│   │   │   ├── vessel.sql             # sqlc queries for vessel
│   │   │   ├── tracking.sql           # sqlc queries for tracking (no UPDATE/DELETE)
│   │   │   └── cargo_events.sql       # sqlc queries for cargo_events
│   │   ├── schema.sql                 # Full schema for sqlc generate
│   │   ├── sqlc.yaml                  # sqlc configuration
│   │   ├── db.go                      # pgxpool.New setup + health check
│   │   ├── cargo_repo.go              # Implements cargo.Repository
│   │   ├── cargo_repo_test.go         # Integration tests (testcontainers)
│   │   ├── vessel_repo.go             # Implements vessel.Repository
│   │   ├── vessel_repo_test.go
│   │   ├── tracking_repo.go           # Implements tracking.Repository
│   │   ├── tracking_repo_test.go
│   │   ├── event_repo.go              # Implements cargo.EventRepository
│   │   └── event_repo_test.go
│   ├── events/                        # Kafka implementation
│   │   ├── producer.go                # Implements cargo.EventPublisher
│   │   └── consumer.go                # Background goroutine — reads topic, writes cargo_events
│   ├── transport/
│   │   └── http/
│   │       ├── cargo_handler.go       # Gin handlers for cargo endpoints
│   │       ├── cargo_handler_test.go  # Handler unit tests (httptest)
│   │       ├── vessel_handler.go
│   │       ├── vessel_handler_test.go
│   │       ├── tracking_handler.go
│   │       ├── tracking_handler_test.go
│   │       ├── dto.go                 # Request/response structs (never sqlc types)
│   │       ├── router.go              # Route registration + middleware chain
│   │       └── errors.go             # Domain error → HTTP status mapping
│   ├── middleware/
│   │   ├── logger.go                  # Zerolog request logging middleware
│   │   ├── recovery.go                # Panic recovery → 500
│   │   └── request_id.go             # Inject request_id into context
│   └── config/
│       └── config.go                  # Config struct + ENV loader
├── pkg/
│   ├── response/
│   │   └── response.go                # JSON success + error response helpers
│   └── validator/
│       └── validator.go               # Input validation helpers
├── Dockerfile                         # Multi-stage build
├── docker-compose.yml                 # api + postgres + kafka + zookeeper
├── Makefile
├── go.mod
├── go.sum
├── README.md
├── .env.example
└── .gitignore
```

---

## 5. Domain Package Pattern (Pandora Style)

Each entity lives in its own package: `internal/domain/<entity>/`.
The **package name IS the entity**. No suffix. Interfaces live alongside models in the same package.

### `internal/domain/cargo/models.go`

```go
// Package cargo contains the cargo domain model and related types.
// This package follows Clean Architecture principles, remaining independent
// of infrastructure and transport concerns.
package cargo

import (
    "time"
    "github.com/google/uuid"
)

// CargoStatus represents the current state of a cargo shipment.
type CargoStatus string

const (
    // CargoStatusPending indicates cargo is registered but not yet in transit.
    CargoStatusPending CargoStatus = "pending"
    // CargoStatusInTransit indicates cargo is currently being transported.
    CargoStatusInTransit CargoStatus = "in_transit"
    // CargoStatusDelivered indicates cargo has been delivered to its destination.
    CargoStatusDelivered CargoStatus = "delivered"
)

// IsValid checks if the status is one of the allowed values.
func (s CargoStatus) IsValid() bool {
    switch s {
    case CargoStatusPending, CargoStatusInTransit, CargoStatusDelivered:
        return true
    default:
        return false
    }
}

// String returns the string representation of CargoStatus.
func (s CargoStatus) String() string { return string(s) }

// Cargo represents a shipment of goods assigned to a vessel.
// This is the pure domain model — never expose sqlc-generated types here.
type Cargo struct {
    ID          uuid.UUID
    Name        string
    Description string
    Weight      float64
    Status      CargoStatus
    VesselID    uuid.UUID
    CreatedAt   time.Time
    UpdatedAt   time.Time
}

// IsDelivered returns true if the cargo has reached its destination.
func (c *Cargo) IsDelivered() bool { return c.Status == CargoStatusDelivered }

// IsInTransit returns true if the cargo is currently being transported.
func (c *Cargo) IsInTransit() bool { return c.Status == CargoStatusInTransit }

// CreateCargoInput contains validated input for creating a new cargo record.
type CreateCargoInput struct {
    Name        string
    Description string
    Weight      float64
    VesselID    uuid.UUID
}
```

### `internal/domain/cargo/errors.go`

```go
// Package cargo contains the cargo domain model and related types.
package cargo

import "errors"

// Domain-level errors for cargo operations.
// These represent business logic failures, not infrastructure failures.
var (
    // ErrNotFound is returned when a cargo cannot be found.
    ErrNotFound = errors.New("cargo not found")
    // ErrInvalidStatus is returned when an invalid status value is provided.
    ErrInvalidStatus = errors.New("invalid cargo status")
    // ErrInvalidInput is returned when request input validation fails.
    ErrInvalidInput = errors.New("invalid input")
)
```

### `internal/domain/cargo/repository.go`

```go
package cargo

import (
    "context"
    "github.com/google/uuid"
)

// Repository defines the interface for cargo data persistence.
// Implemented by internal/postgres/cargo_repo.go.
// Following Clean Architecture: domain defines the contract, infrastructure implements it.
type Repository interface {
    // Create persists a new cargo record and returns the created entity.
    // Returns error if vessel does not exist or DB operation fails.
    Create(ctx context.Context, input CreateCargoInput) (*Cargo, error)

    // GetByID retrieves a cargo by its unique ID.
    // Returns ErrNotFound if cargo does not exist.
    GetByID(ctx context.Context, id uuid.UUID) (*Cargo, error)

    // List retrieves all cargo records. Returns empty slice if none found.
    List(ctx context.Context) ([]*Cargo, error)

    // UpdateStatus updates the cargo status field.
    // Returns ErrNotFound if cargo does not exist.
    UpdateStatus(ctx context.Context, id uuid.UUID, status CargoStatus) (*Cargo, error)
}
```

### `internal/domain/cargo/service.go`

```go
package cargo

import (
    "context"
    "github.com/google/uuid"
)

// Service defines the interface for cargo business logic.
// Implemented by internal/service/cargo_service.go.
type Service interface {
    // CreateCargo registers a new cargo shipment assigned to a vessel.
    // Returns ErrInvalidInput if validation fails.
    // Returns vessel.ErrNotFound if the assigned vessel does not exist.
    CreateCargo(ctx context.Context, input CreateCargoInput) (*Cargo, error)

    // GetCargo retrieves a cargo by ID.
    // Returns ErrNotFound if cargo does not exist.
    GetCargo(ctx context.Context, id uuid.UUID) (*Cargo, error)

    // ListCargoes retrieves all cargo records.
    ListCargoes(ctx context.Context) ([]*Cargo, error)

    // UpdateCargoStatus transitions cargo to a new status.
    // Appends an immutable tracking entry on every status change.
    // Emits a Kafka cargo.status_changed event after successful DB write.
    // Returns ErrNotFound if cargo does not exist.
    // Returns ErrInvalidStatus if the provided status is not a valid CargoStatus.
    UpdateCargoStatus(ctx context.Context, id uuid.UUID, status CargoStatus) (*Cargo, error)
}
```

### `internal/domain/cargo/events.go`

```go
package cargo

import (
    "context"
    "time"
    "github.com/google/uuid"
)

// CargoEvent is an immutable record of a cargo status transition.
// Written to the cargo_events table by the Kafka consumer worker.
type CargoEvent struct {
    ID        uuid.UUID
    CargoID   uuid.UUID
    OldStatus CargoStatus
    NewStatus CargoStatus
    Timestamp time.Time
}

// StatusChangedEvent is the Kafka message payload emitted on every status change.
type StatusChangedEvent struct {
    ID        string      `json:"id"`
    EventType string      `json:"event_type"`
    CargoID   string      `json:"cargo_id"`
    OldStatus CargoStatus `json:"old_status"`
    NewStatus CargoStatus `json:"new_status"`
    Timestamp time.Time   `json:"timestamp"`
}

// EventPublisher defines the contract for publishing cargo domain events.
// Implemented by internal/events/producer.go (Kafka).
// Interface keeps domain decoupled from Kafka specifics.
type EventPublisher interface {
    // PublishStatusChanged emits a cargo status change event to Kafka.
    // Must be called after a successful DB write, never before.
    // Errors are logged but MUST NOT fail the HTTP request (fire-and-forget).
    PublishStatusChanged(ctx context.Context, event StatusChangedEvent) error
}

// EventRepository defines the contract for persisting cargo events.
// Implemented by internal/postgres/event_repo.go.
type EventRepository interface {
    // Store persists a cargo event record. Append-only — never updated or deleted.
    Store(ctx context.Context, event CargoEvent) error

    // ListByCargoID retrieves all events for a cargo in chronological order.
    ListByCargoID(ctx context.Context, cargoID uuid.UUID) ([]*CargoEvent, error)
}
```

---

## 6. Database Schema

### `vessels` table

| Column | Type | Notes |
|---|---|---|
| id | UUID PK | `gen_random_uuid()` |
| name | text NOT NULL | |
| capacity | numeric NOT NULL | max cargo weight |
| current_location | text NOT NULL | last known location |
| created_at | timestamptz | `default now()` |
| updated_at | timestamptz | updated on change |

### `cargoes` table

| Column | Type | Notes |
|---|---|---|
| id | UUID PK | `gen_random_uuid()` |
| name | text NOT NULL | |
| description | text | optional |
| weight | numeric NOT NULL | |
| status | text NOT NULL | `pending` / `in_transit` / `delivered` |
| vessel_id | UUID FK | `REFERENCES vessels(id)` |
| created_at | timestamptz | `default now()` |
| updated_at | timestamptz | |

### `tracking_entries` table — **APPEND-ONLY. NEVER UPDATE OR DELETE.**

| Column | Type | Notes |
|---|---|---|
| id | UUID PK | `gen_random_uuid()` |
| cargo_id | UUID FK | `REFERENCES cargoes(id)` |
| location | text NOT NULL | |
| status | text NOT NULL | cargo status at this moment |
| note | text | optional |
| timestamp | timestamptz | when recorded |

### `cargo_events` table — **Written by Kafka consumer. Append-only.**

| Column | Type | Notes |
|---|---|---|
| id | UUID PK | `gen_random_uuid()` |
| cargo_id | UUID FK | `REFERENCES cargoes(id)` |
| old_status | text NOT NULL | |
| new_status | text NOT NULL | |
| timestamp | timestamptz | |

### Indexes

```sql
CREATE INDEX idx_cargoes_vessel_id          ON cargoes(vessel_id);
CREATE INDEX idx_cargoes_status             ON cargoes(status);
CREATE INDEX idx_tracking_entries_cargo_id  ON tracking_entries(cargo_id);
CREATE INDEX idx_tracking_entries_timestamp ON tracking_entries(timestamp);
CREATE INDEX idx_cargo_events_cargo_id      ON cargo_events(cargo_id);
```

---

## 7. Service Startup Sequence (Single Binary)

```
1.  Load config from ENV
2.  Init Zerolog logger
3.  Connect PostgreSQL (pgxpool.New)
4.  Run pending migrations (golang-migrate)
5.  Init sqlc queries
6.  Init Kafka producer
7.  Start Kafka consumer goroutine (background, listens to cargo-status-changes)
8.  Wire DI: repositories → services → handlers
9.  Register Gin router + middleware
10. Register /health + /ready endpoints
11. Start HTTP server on :8080
12. Graceful shutdown on SIGTERM/SIGINT
13. Wait for in-flight requests + Kafka consumer flush
```

---

## 8. Kafka Event Flow (In-Process Consumer)

```
┌─────────────────────────────────────────────────────────────────────┐
│                 KAFKA EVENT FLOW (Single Binary)                     │
├─────────────────────────────────────────────────────────────────────┤
│                                                                      │
│  1. PATCH /api/v1/cargoes/:id/status                                │
│                        │                                            │
│                        ▼                                            │
│  2. cargoService.UpdateCargoStatus()                                │
│     • Updates cargoes table                                         │
│     • Appends to tracking_entries (append-only)                     │
│     • Calls eventPublisher.PublishStatusChanged()                   │
│                        │                                            │
│                        ▼                                            │
│  3. Kafka Producer → "cargo-status-changes" topic                   │
│     (fire-and-forget: errors logged, never fail the request)        │
│                        │                                            │
│                        ▼                                            │
│  4. Kafka Consumer goroutine (same process) reads message           │
│                        │                                            │
│                        ▼                                            │
│  5. Consumer writes CargoEvent to cargo_events table (append-only)  │
│                                                                      │
└─────────────────────────────────────────────────────────────────────┘
```

---

## 9. REST API Endpoints

| Method | Path | Description |
|---|---|---|
| POST | `/api/v1/cargoes` | Create cargo |
| GET | `/api/v1/cargoes` | List all cargoes |
| GET | `/api/v1/cargoes/:id` | Get cargo by ID |
| PATCH | `/api/v1/cargoes/:id/status` | Update cargo status → emits Kafka event |
| POST | `/api/v1/vessels` | Create vessel |
| GET | `/api/v1/vessels` | List all vessels |
| GET | `/api/v1/vessels/:id` | Get vessel by ID |
| POST | `/api/v1/cargoes/:id/tracking` | Add tracking entry |
| GET | `/api/v1/cargoes/:id/tracking` | Get full tracking history |
| GET | `/health` | Liveness check |
| GET | `/ready` | Readiness check (DB + Kafka) |

---

## 10. JSON Response Contract

**Success response:**

```json
{
  "data": { "id": "550e8400-e29b-41d4-a716-446655440000" },
  "meta": {
    "request_id": "550e8400-e29b-41d4-a716-446655440000"
  }
}
```

**Error response:**

```json
{
  "error": {
    "code": "CARGO_NOT_FOUND",
    "message": "cargo not found",
    "request_id": "550e8400-e29b-41d4-a716-446655440000"
  }
}
```

---

## 11. HTTP Error Mapping

| Domain Error | HTTP | Code |
|---|---|---|
| `cargo.ErrNotFound` | 404 | `CARGO_NOT_FOUND` |
| `vessel.ErrNotFound` | 404 | `VESSEL_NOT_FOUND` |
| `cargo.ErrInvalidStatus` | 400 | `INVALID_STATUS` |
| `cargo.ErrInvalidInput` / `vessel.ErrInvalidInput` | 400 | `INVALID_INPUT` |
| `vessel.ErrCapacityExceeded` | 422 | `CAPACITY_EXCEEDED` |
| any unhandled error | 500 | `INTERNAL_ERROR` |

---

## 12. Docker Compose Services

```yaml
services:
  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_USER: deus
      POSTGRES_PASSWORD: deus
      POSTGRES_DB: deus
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U deus"]
      interval: 5s
      timeout: 5s
      retries: 5

  zookeeper:
    image: confluentinc/cp-zookeeper:latest
    environment:
      ZOOKEEPER_CLIENT_PORT: 2181
    ports:
      - "2181:2181"

  kafka:
    image: confluentinc/cp-kafka:latest
    depends_on:
      - zookeeper
    environment:
      KAFKA_BROKER_ID: 1
      KAFKA_ZOOKEEPER_CONNECT: zookeeper:2181
      KAFKA_ADVERTISED_LISTENERS: PLAINTEXT://kafka:9092
      KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR: 1
      KAFKA_AUTO_CREATE_TOPICS_ENABLE: "true"
    ports:
      - "9092:9092"
    healthcheck:
      test: ["CMD-SHELL", "kafka-topics --bootstrap-server localhost:9092 --list"]
      interval: 10s
      timeout: 10s
      retries: 5

  api:
    build:
      context: .
      dockerfile: Dockerfile
    ports:
      - "8080:8080"
    depends_on:
      postgres:
        condition: service_healthy
      kafka:
        condition: service_healthy
    environment:
      DB_URL: postgres://deus:deus@postgres:5432/deus?sslmode=disable
      KAFKA_BROKERS: kafka:9092
      SERVER_PORT: 8080
      LOG_LEVEL: info
    restart: unless-stopped

volumes:
  postgres_data:
```

---

## 13. Testing Strategy

### Unit Tests (Table-Driven with Mocks)

```go
func TestCargoService_UpdateCargoStatus(t *testing.T) {
    cargoID := uuid.New()

    tests := []struct {
        name      string
        cargoID   uuid.UUID
        newStatus cargo.CargoStatus
        mockSetup func(*mocks.MockCargoRepository, *mocks.MockEventPublisher)
        wantErr   bool
        wantErrIs error
    }{
        {
            // Given: a valid cargo in pending status
            // When: UpdateCargoStatus called with in_transit
            // Then: returns updated cargo, no error
            name:      "success: pending → in_transit",
            cargoID:   cargoID,
            newStatus: cargo.CargoStatusInTransit,
            mockSetup: func(r *mocks.MockCargoRepository, p *mocks.MockEventPublisher) {
                r.On("GetByID", mock.Anything, cargoID).Return(&cargo.Cargo{
                    ID:     cargoID,
                    Status: cargo.CargoStatusPending,
                }, nil)
                r.On("UpdateStatus", mock.Anything, cargoID, cargo.CargoStatusInTransit).
                    Return(&cargo.Cargo{ID: cargoID, Status: cargo.CargoStatusInTransit}, nil)
                p.On("PublishStatusChanged", mock.Anything, mock.Anything).Return(nil)
            },
            wantErr: false,
        },
        {
            // Given: a non-existent cargo
            // When: UpdateCargoStatus called
            // Then: returns ErrNotFound
            name:      "error: cargo not found",
            cargoID:   uuid.New(),
            newStatus: cargo.CargoStatusInTransit,
            mockSetup: func(r *mocks.MockCargoRepository, p *mocks.MockEventPublisher) {
                r.On("GetByID", mock.Anything, mock.Anything).Return(nil, cargo.ErrNotFound)
            },
            wantErr:   true,
            wantErrIs: cargo.ErrNotFound,
        },
        {
            // Given: an invalid status string
            // When: UpdateCargoStatus called
            // Then: returns ErrInvalidStatus, no DB call made
            name:      "error: invalid status",
            cargoID:   cargoID,
            newStatus: cargo.CargoStatus("bad_status"),
            mockSetup: func(r *mocks.MockCargoRepository, p *mocks.MockEventPublisher) {},
            wantErr:   true,
            wantErrIs: cargo.ErrInvalidStatus,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mockRepo := new(mocks.MockCargoRepository)
            mockPub := new(mocks.MockEventPublisher)
            tt.mockSetup(mockRepo, mockPub)

            svc := service.NewCargoService(mockRepo, mockPub)
            _, err := svc.UpdateCargoStatus(context.Background(), tt.cargoID, tt.newStatus)

            if tt.wantErr {
                require.Error(t, err)
                if tt.wantErrIs != nil {
                    assert.ErrorIs(t, err, tt.wantErrIs)
                }
            } else {
                require.NoError(t, err)
            }

            mockRepo.AssertExpectations(t)
            mockPub.AssertExpectations(t)
        })
    }
}
```

### Integration Tests (Testcontainers)

```go
func TestCargoRepository_Integration(t *testing.T) {
    // Given: a running PostgreSQL container
    ctx := context.Background()
    container, db, err := startPostgresContainer(ctx)
    require.NoError(t, err)
    defer container.Terminate(ctx)

    runMigrations(t, db)
    repo := postgres.NewCargoRepository(db)

    t.Run("Create and GetByID", func(t *testing.T) {
        // When: a cargo is created
        input := cargo.CreateCargoInput{
            Name:     "Test Cargo",
            Weight:   100.0,
            VesselID: createTestVessel(t, db),
        }
        created, err := repo.Create(ctx, input)

        // Then: it can be retrieved by ID
        require.NoError(t, err)
        require.NotNil(t, created)

        found, err := repo.GetByID(ctx, created.ID)
        require.NoError(t, err)
        assert.Equal(t, created.ID, found.ID)
        assert.Equal(t, cargo.CargoStatusPending, found.Status)
    })
}
```

---

## 14. Hard Rules (Enforced by AI)

| Rule | Detail |
|---|---|
| ❌ No business logic in handlers | `bind → validate → call service → respond` only |
| ❌ No raw SQL in Go files | sqlc ONLY — all queries in `.sql` files |
| ❌ No cross-layer imports | Domain never imports transport, postgres, or events |
| ❌ No sqlc structs in responses | Always map to domain structs or DTOs |
| ❌ No `panic()` | Error returns always |
| ❌ No `fmt.Println()` | Zerolog only |
| ❌ No `context.TODO()` | Propagate context always |
| ✅ GoDoc on every exported symbol | No exceptions |
| ✅ Tests FIRST — TDD | Failing tests before any implementation |
| ✅ Table-driven tests with Given/When/Then | All unit tests |
| ✅ `IsValid()` + `String()` on all enums | `CargoStatus` required |
| ✅ Pointer returns `*Cargo`, `*Vessel` | Consistent with Pandora pattern |
| ✅ `tracking_entries` append-only | No `UPDATE`/`DELETE` in `tracking.sql` |
| ✅ Single binary with in-process consumer | Consumer goroutine started in `main.go` |

---

## 15. Architecture Decision Records (ADRs)

| # | Decision | Rationale |
|---|---|---|
| ADR-001 | sqlc over GORM | Type-safe, compile-time verified, no magic, idiomatic Go |
| ADR-002 | Domain-per-entity packages | Co-location of models + interfaces = easier navigation + true bounded contexts |
| ADR-003 | In-process Kafka consumer | Single binary simplifies deployment for this challenge scope |
| ADR-004 | Append-only tracking entries | Full auditability — every state transition is permanently recorded |
| ADR-005 | Fire-and-forget Kafka publish | Kafka failure must never degrade the HTTP API response |
| ADR-006 | Testcontainers for integration tests | Reproducible, no external dependency, runs in CI |
| ADR-007 | golang-migrate with raw SQL | Explicit, reviewable, no ORM magic in schema management |

---

## ✅ Final Rule

**If something is unclear → ask before coding.
All code MUST comply with this architecture. No deviation without explicit approval.**

---
