# DEUS Logistics API

A production-grade backend service for a logistics company that ships goods via vessels.

**Status:** Bootstrapping (Task #1 in progress)

---

## Quick Start

### Prerequisites
- Go 1.21+
- Docker & Docker Compose
- PostgreSQL 15
- Kafka (via docker-compose)

### Setup

```bash
# Install dependencies
go mod download

# Install build tools
make install-tools

# Start services (postgres + kafka)
make docker-up

# Run migrations
make migrate-up

# Build and run API
make run
```

Server starts on `http://localhost:8080`

---

## Architecture

This project follows **Clean Architecture** with:
- **Domain-per-entity** packages (Pandora Exchange pattern)
- **sqlc** for type-safe, compile-time-verified database access
- **Single binary** API + in-process Kafka consumer goroutine
- **Table-driven tests** with mockgen mocks + testcontainers

See [.copilot/ARCHITECTURE.md](.copilot/ARCHITECTURE.md) for detailed specification.

---

## Tech Stack

| Category | Technology |
|---|---|
| Language | Go 1.21+ |
| HTTP Framework | Gin |
| Database | PostgreSQL 15 |
| Data Access | sqlc (SQL-first) |
| Migrations | golang-migrate |
| Event Bus | Kafka (segmentio/kafka-go) |
| Logging | Zerolog |
| Testing | Go testing + testify + mockgen + testcontainers |
| Containers | Docker + docker-compose |

---

## Folder Structure

```
/deus-logistics-api
├── cmd/
│   └── api/
│       └── main.go                    # Single binary — API + consumer goroutine
├── internal/
│   ├── domain/
│   │   ├── cargo/                     # Cargo bounded context
│   │   ├── vessel/                    # Vessel bounded context
│   │   └── tracking/                  # Tracking bounded context
│   ├── service/                       # Service implementations + unit tests
│   ├── postgres/                      # sqlc implementations + integration tests
│   │   ├── migrations/
│   │   ├── queries/
│   │   └── sqlc.yaml
│   ├── events/                        # Kafka producer + consumer
│   ├── transport/
│   │   └── http/                      # Gin handlers + DTOs
│   ├── middleware/                    # Logging, recovery, request_id middleware
│   └── config/                        # Config struct + ENV loader
├── pkg/
│   ├── response/                      # JSON response helpers
│   └── validator/                     # Input validation
├── .copilot/
│   ├── ARCHITECTURE.md                # Single source of truth
│   ├── system_prompt.md               # Role and behavior
│   ├── rules.md                       # Workflow patterns
│   └── config.json                    # Linters and banned patterns
├── Dockerfile
├── docker-compose.yml
├── Makefile
├── go.mod
├── go.sum
├── README.md (this file)
├── .env.example
└── .gitignore
```

---

## Master Task List

| # | Task | Owner | Status | Priority | Estimate | Details |
|---|---|---|---|---|---|---|
| 1 | Bootstrap repo skeleton | Copilot | done | P0 | 2h | go.mod, .gitignore, Makefile, folder structure, README |
| 2 | Database schema + sqlc config | Copilot | done | P0 | 3h | schema.sql, sqlc.yaml, queries for all entities |
| 3 | Migrations (up/down) | Copilot | done | P0 | 2h | 000001_init.up.sql + down.sql via golang-migrate |
| 4 | Domain layer | Copilot | done | P0 | 3h | Structs, interfaces, sentinel errors (cargo, vessel, tracking) |
| 5 | Repository implementations | Copilot | done | P0 | 4h | sqlc-backed impls for all repos in /internal/postgres/ |
| 6 | Service implementations + unit tests | Copilot | done | P0 | 5h | cargo_service.go, vessel_service.go, tracking_service.go (with mocks) |
| 7 | Kafka producer | Copilot | done | P1 | 2h | EventPublisher implementation in /internal/events/producer.go |
| 8 | Kafka consumer (in-process) | Copilot | done | P1 | 2h | Consumer goroutine in main, writes to cargo_events table |
| 9 | Gin router + handlers + DTOs | Copilot | done | P0 | 4h | All endpoints from ARCHITECTURE.md |
| 10 | Error handling + middleware | Copilot | done | P0 | 3h | Domain error → HTTP status mapping + request logging |
| 11 | Config + startup sequence | Copilot | done | P0 | 2h | Config struct, main.go startup sequence (13 steps) |
| 12 | Dockerfile (multi-stage) | Copilot | done | P0 | 1h | Single image for API + consumer |
| 13 | docker-compose.yml | Copilot | done | P0 | 1h | api + postgres + kafka + zookeeper |
| 14 | Health + readiness endpoints | Copilot | done | P1 | 1h | /health + /ready |
| 15 | Integration tests | Copilot | done | P0 | 4h | All repos tested against real Postgres + Kafka |
| 16 | README final + API docs | Copilot | done | P0 | 2h | Full setup guide + all endpoint examples + request/response samples |

---

## Development

### Running Tests

```bash
# All tests
make test

# With coverage
make test-coverage

# Specific package
go test -v ./internal/service/...
```

### Database Migrations

```bash
# Apply pending migrations
make migrate-up

# Rollback last migration
make migrate-down
```

### Rebuilding sqlc Generated Code

```bash
# After updating queries/*.sql
make sqlc
```

### Generating Mocks

```bash
# After updating an interface
mockgen -source=internal/domain/cargo/repository.go -destination=internal/service/mocks/mock_cargo_repository.go -package=mocks
```

---

## REST API Endpoints

### Core Endpoints

| Method | Path | Description | Status |
|---|---|---|---|
| POST | `/api/v1/cargoes` | Create cargo | ✅ Implemented |
| GET | `/api/v1/cargoes` | List all cargoes | ✅ Implemented |
| GET | `/api/v1/cargoes/:id` | Get cargo by ID | ✅ Implemented |
| PATCH | `/api/v1/cargoes/:id/status` | Update cargo status → Kafka event + tracking | ✅ Implemented |
| POST | `/api/v1/vessels` | Create vessel | ✅ Implemented |
| GET | `/api/v1/vessels` | List all vessels | ✅ Implemented |
| GET | `/api/v1/vessels/:id` | Get vessel by ID | ✅ Implemented |
| PATCH | `/api/v1/vessels/:id/location` | Update vessel location | ✅ Implemented |
| POST | `/api/v1/cargoes/:id/tracking` | Add tracking entry (append-only) | ✅ Implemented |
| GET | `/api/v1/cargoes/:id/tracking` | Get tracking history | ✅ Implemented |
| GET | `/health` | Liveness check | ✅ Implemented |
| GET | `/ready` | Readiness check | ✅ Implemented |

### API Examples

#### Create Cargo

```bash
curl -X POST http://localhost:8080/api/v1/cargoes \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Electronics Shipment",
    "description": "Fragile electronics for Europe",
    "weight": 500.0,
    "vessel_id": "550e8400-e29b-41d4-a716-446655440000"
  }'
```

**Response (201 Created):**
```json
{
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440001",
    "name": "Electronics Shipment",
    "description": "Fragile electronics for Europe",
    "weight": 500.0,
    "status": "pending",
    "vessel_id": "550e8400-e29b-41d4-a716-446655440000",
    "created_at": "2026-03-24T10:30:00Z",
    "updated_at": "2026-03-24T10:30:00Z"
  },
  "meta": {
    "request_id": "550e8400-e29b-41d4-a716-446655440002"
  }
}
```

#### Update Cargo Status (triggers Kafka event + tracking)

```bash
curl -X PATCH http://localhost:8080/api/v1/cargoes/550e8400-e29b-41d4-a716-446655440001/status \
  -H "Content-Type: application/json" \
  -d '{
    "status": "in_transit"
  }'
```

**Response (200 OK):**
```json
{
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440001",
    "name": "Electronics Shipment",
    "status": "in_transit",
    "vessel_id": "550e8400-e29b-41d4-a716-446655440000",
    "updated_at": "2026-03-24T10:35:00Z"
  },
  "meta": {
    "request_id": "550e8400-e29b-41d4-a716-446655440003"
  }
}
```

Side effects:
- ✅ Cargo status updated in `cargoes` table
- ✅ Tracking entry appended to `tracking_entries` table
- ✅ Kafka event published to `cargo-status-changes` topic
- ✅ Kafka consumer persists event to `cargo_events` table (append-only)

#### Get Tracking History

```bash
curl http://localhost:8080/api/v1/cargoes/550e8400-e29b-41d4-a716-446655440001/tracking
```

**Response (200 OK):**
```json
{
  "data": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440004",
      "cargo_id": "550e8400-e29b-41d4-a716-446655440001",
      "location": "Unknown",
      "status": "in_transit",
      "note": "Status changed from pending to in_transit",
      "timestamp": "2026-03-24T10:35:00Z"
    }
  ],
  "meta": {
    "request_id": "550e8400-e29b-41d4-a716-446655440005"
  }
}
```

#### Error Response

```bash
curl http://localhost:8080/api/v1/cargoes/invalid-id
```

**Response (400 Bad Request):**
```json
{
  "error": {
    "code": "INVALID_INPUT",
    "message": "invalid cargo id",
    "request_id": "550e8400-e29b-41d4-a716-446655440006"
  }
}
```

#### Cargo Not Found

```bash
curl http://localhost:8080/api/v1/cargoes/550e8400-0000-0000-0000-000000000000
```

**Response (404 Not Found):**
```json
{
  "error": {
    "code": "CARGO_NOT_FOUND",
    "message": "cargo not found",
    "request_id": "550e8400-e29b-41d4-a716-446655440007"
  }
}
```

---

## JSON Response Contract

### Success Response

```json
{
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "Electronics",
    "status": "pending"
  },
  "meta": {
    "request_id": "550e8400-e29b-41d4-a716-446655440001"
  }
}
```

### Error Response

```json
{
  "error": {
    "code": "CARGO_NOT_FOUND",
    "message": "cargo not found",
    "request_id": "550e8400-e29b-41d4-a716-446655440001"
  }
}
```

### HTTP Status Codes

- `200 OK` — Successful GET
- `201 Created` — Successful POST
- `400 Bad Request` — Invalid input / bad status
- `404 Not Found` — Cargo or vessel not found
- `422 Unprocessable Entity` — Business rule violation (e.g., capacity exceeded)
- `500 Internal Server Error` — Unexpected server error

---

## Configuration

Configuration is via environment variables. Copy `.env.example` to `.env`:

```bash
cp .env.example .env
```

Then edit `.env` with your values. Required variables:
- `DB_HOST` — PostgreSQL host
- `DB_PORT` — PostgreSQL port
- `DB_USER` — PostgreSQL user
- `DB_PASSWORD` — PostgreSQL password
- `DB_NAME` — PostgreSQL database name
- `KAFKA_BROKER` — Kafka broker address
- `SERVER_PORT` — API port (default: 8080)

---

## Named Branches & Commits

Each task uses conventional commits:

```
feature/<task-name>
feat: <description>
fix: <description>
test: <description>
chore: <description>
ci: <description>
```

---

## Rules & Constraints

See [.copilot/ARCHITECTURE.md](.copilot/ARCHITECTURE.md), [.copilot/system_prompt.md](.copilot/system_prompt.md), and [.copilot/rules.md](.copilot/rules.md) for:
- Complete architecture specification
- Mandatory patterns (sqlc, Zerolog, Clean Architecture)
- Banned patterns (panic, fmt.Println, GORM, raw SQL)
- TDD workflow
- Coding standards

---

## License

MIT License

Copyright (c) 2025 Alex Necsoiu

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.