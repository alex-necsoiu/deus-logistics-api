# ✅ Pandora Exchange — Backend Architecture Specification (v1.0)

> **Status:** Approved  
> **Scope:** Backend architecture — Phase 1 (User Service)  
> **Audience:** Backend engineers + AI coding assistants  
> **Rule:** This is the **single source of truth** for backend development. ALL code must follow this spec.

---

## 1. Overview

Pandora Exchange is a high-performance digital asset trading platform focused on:

- Low latency execution
- High security cryptography and custody
- Regulatory compliance (KYC/AML/audit logs)
- Deterministic and auditable backend state
- Service isolation and scalability

This document defines:

- Architecture & service boundaries
- Domain interfaces
- Event system
- DB schema and access pattern
- Environments & CI/CD
- Security and compliance rules
- Development and testing requirements
- Rules for human + AI coding

All development MUST strictly follow this architecture.

---

## 2. Architectural Goals

| Goal | Description |
|---|---|
Low latency | Optimized for high I/O trading environments |
Security | Argon2id, JWT, Vault, HSM, TLS |
Auditability | Immutable audit logs, KYC/AML flows |
Reliability | Distributed, fault tolerant, event-driven |
Scalability | Kubernetes, horizontal scaling |
Determinism | SQL-first (`sqlc`), strict types, explicit contracts |
Maintainability | Clean architecture, modular services |

---

## 3. Technology Stack

| Category | Technology |
|---|---|
Language | Go ≥ 1.24 |
Framework | Gin (REST) |
Internal RPC | gRPC |
Database | PostgreSQL |
Data Access | sqlc (SQL-first) |
Async Bus | Redis Streams (Kafka future migration) |
Cache | Redis |
Auth | go-jwt, Argon2id |
Logging | Zerolog |
Tracing | OpenTelemetry |
Secrets | Hashicorp Vault |
Containers | Docker |
Orchestration | Kubernetes |
CI/CD | GitHub Actions |
Config | ENV + YAML config per env |

---

## 4. Microservices Responsibilities

| Service | Responsibilities |
|---|---|
User Service *(Phase 1)* | Auth, JWT, refresh tokens, KYC events |
Wallet Service | Hot/cold wallet mgmt, ERC-4337 |
Trading Engine *(future)* | Matching engine, order book |
Payments & Settlements | Fiat & crypto settlements |
Notification Service | WebSockets & async messages |
Admin & Compliance | AML/KYC review, audit logs |

> **Initial phase scope: User Service only**

---

## 5. Communication Model

| Path | Protocol |
|---|---|
External clients | HTTP REST (Gin) |
Internal services | gRPC |
Async events | Redis Streams |
Tracing | OpenTelemetry |

### Event Envelope

```json
{
  "id": "uuid",
  "timestamp": "ISO8601",
  "trace_id": "otel_trace_id",
  "event_type": "user.created",
  "payload": {}
}
```

---

## 6. Code Structure Per Service

```
/service-name
  /cmd/service
    main.go
  /internal
    /domain
    /repository
    /postgres        # sqlc layer
    /transport
       /http         # Gin
       /grpc         # gRPC
    /events
    /config
    /middleware
    /observability
  /pkg               # optional shared libs
  go.mod
  Makefile
```

### Rules

- ❌ No business logic in handlers
- ✅ Domain never imports infrastructure
- ✅ sqlc never exposed to API layer
- ✅ No cross-service DB access

---

## 7. Domain Interfaces (User Service)

```go
type UserService interface {
  Register(ctx, email, password string) (User, error)
  Login(ctx, email, password string) (TokenPair, error)
  GetByID(ctx, id string) (User, error)
  UpdateKYC(ctx, id, status string) error
}
```

```go
type UserRepository interface {
  Create(ctx context.Context, params CreateUserParams) (User, error)
  GetByEmail(ctx context.Context, email string) (User, error)
  GetByID(ctx context.Context, id string) (User, error)
  UpdateKYC(ctx context.Context, id, status string) error
  SoftDelete(ctx context.Context, id string) error
}
```

---

## 8. Database Schema

### `users` table

| Column | Type | Notes |
|---|---|---|
id | UUID PK | unique user id |
email | text unique | required |
full_name | text | optional |
hashed_password | text | Argon2id |
kyc_status | text | pending / verified / rejected |
created_at / updated_at / deleted_at | timestamps | soft delete |

### `refresh_tokens` table

| Column | Type |
|---|---|
token | PK |
user_id | FK |
expires_at | timestamp |
created_at | timestamp |

> **Explicit migrations required**

---

## 9. Environments

| Env | Purpose | Data |
|---|---|---|
dev | Local | seeded mock data |
sandbox | QA | synthetic live-like data |
audit | compliance | anonymized prod clone |
prod | live | real secure data |

`APP_ENV = dev | sandbox | audit | prod`

---

## 10. CI/CD

| Stage | Action |
|---|---|
PR | linter, tests, sqlc build |
main merge | build + deploy to sandbox |
nightly | security scan + migrations test |
prod deploy | manual gate |

---

## 11. Security Standards

- Argon2id hashing
- JWT access + refresh tokens
- Vault secrets (no secrets in code/env)
- TLS everywhere
- HSM for wallet operations in production
- No sensitive data in logs
- Immutable audit logs

---

## 12. Make Targets

```
make dev-up
make dev-down
make migrate
make sqlc
make test
```

---

## 13. Service Startup Sequence

1. Load config  
2. Init logging  
3. Connect PostgreSQL + sqlc  
4. Connect Redis + Streams  
5. Start OTEL tracing  
6. Start gRPC server  
7. Start REST server  
8. Register health checks  
9. Serve requests  

---

## 14. AI Coding Rules

AI MUST generate:

- Folder structure
- go.mod + Makefile
- sqlc schema + queries
- migrations
- Gin handlers + DTOs
- gRPC proto + stubs
- Redis Streams handlers
- unit tests + mocks
- config files
- Dockerfile + docker compose
- OTEL instrumentation

AI MUST NOT:

- Mix domain logic with transport layer
- Access DB directly from handlers
- Expose sqlc structs to API
- Skip tests or comments

## 15 Error Handling

You must now implement the Pandora Exchange error-handling system.

Follow the Pandora error strategy:

- Use sentinel domain errors (ErrUserExists, ErrInvalidInput, etc.)
- Create AppError struct with Code, Message, TraceID, HTTPStatus
- Provide NewAppError constructor
- Create Gin middleware to map domain errors → HTTP responses
- Create gRPC interceptor to map domain errors → status codes
- Attach OTEL trace IDs to all error responses
- Use Zerolog for logging errors (no sensitive data)
- Do NOT expose internal errors to clients
- Use table-driven tests for error mapping
- File paths MUST follow our architecture layout
- Add GoDoc comments on all functions
- Tests FIRST, then implementation

Output patch plan with:

- Branch name
- Commit messages
- File map (full paths)
- Tests first, failing
- Code after tests
- Commands to run tests
- Expected outputs


## ✅ Final Rule

**If something is unclear → ask before coding.  
All code MUST comply with this architecture.**

---

# End of Architecture Document
