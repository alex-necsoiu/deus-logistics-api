# Application Layer - Use Case Architecture

## Overview

This package implements the **Application Layer** following Clean Architecture principles, organiz the business logic orchestration layer between domain models and HTTP handlers.

## Design Principles

### 1. **Single Responsibility (Use Cases)**
Each use case is a focused struct that performs ONE operation:
- `CreateCargoUseCase` - Only creates new cargo
- `GetCargoUseCase` - Only retrieves a cargo by ID
- `ListCargosUseCase` - Only lists all cargos
- `ListCargosByVesselUseCase` - Only lists cargos for a vessel
- `UpdateCargoStatusUseCase` - Only updates cargo status with full orchestration

### 2. **Dependency Injection via Interfaces**
All dependencies are injected through constructor functions:
```go
func NewUpdateCargoStatusUseCase(
    cargoRepo CargoRepository,
    trackingRepo TrackingRepository,
    publisher EventPublisher,
) *UpdateCargoStatusUseCase
```

Benefits:
- Enables testing with mock implementations
- Decouples from concrete implementations
- Easy to swap implementations

### 3. **Separation of Concerns**
- **Domain Layer**: Business rules (state machine transitions, validation)
- **Application Layer**: Orchestration (calling domain methods, persistence, side effects)
- **Transport Layer**: HTTP concerns (request/response handling)

### 4. **Pure Orchestration**
Use cases are pure orchestrators - they DO NOT contain business logic:
- ✅ Calls domain methods for validation
- ✅ Manages persistence transactions
- ✅ Coordinates side effects (Kafka, tracking)
- ❌ Do NOT contain business rules
- ❌ Do NOT validate business constraints (domain does that)

Example from `UpdateCargoStatusUseCase.Execute()`:
```go
// Domain enforces all business rules
if err := currentCargo.UpdateStatus(newStatus); err != nil {
    return nil, err  // Invalid transition, domain rejected it
}

// Persistence
updatedCargo, err := uc.cargoRepo.UpdateStatus(ctx, id, newStatus)

// Side effects: fire-and-forget pattern
_ = uc.publisher.PublishStatusChanged(ctx, event)
```

## File Structure

```
internal/application/cargo/
├── interfaces.go          # Repository and publisher interfaces
├── create_cargo.go        # CreateCargo use case
├── get_cargo.go           # GetCargo use case
├── list_cargos.go         # ListCargos use case
├── list_by_vessel.go      # ListCargosByVessel use case
├── update_status.go       # UpdateCargoStatus use case
└── manager.go             # CargoApplicationManager (wires use cases)
```

## Use Case Pattern

Each use case follows this pattern:

```go
type CreateCargoUseCase struct {
    repo CargoRepository
}

func NewCreateCargoUseCase(repo CargoRepository) *CreateCargoUseCase {
    return &CreateCargoUseCase{repo: repo}
}

func (uc *CreateCargoUseCase) Execute(ctx context.Context, input cargo.CreateCargoInput) (*cargo.Cargo, error) {
    // 1. Validate input (or delegate to domain)
    if err := input.Validate(); err != nil {
        return nil, err
    }

    // 2. Call repository/domain
    result, err := uc.repo.Create(ctx, input)
    if err != nil {
        return nil, fmt.Errorf("create_cargo: %w", err)
    }

    // 3. Handle side effects
    zerolog.Ctx(ctx).Info().Msg("cargo created")

    return result, nil
}
```

## Integration with HTTP Layer

The HTTP handlers use the application manager to access use cases:

```go
// Handler receives manager in dependency injection
type CargoHandler struct {
    app *appcargo.CargoApplicationManager
}

// Use cases are called for orchestration
func (h *CargoHandler) CreateCargo(c *gin.Context) {
    result, err := h.app.CreateCargo.Execute(ctx, input)
    // Handle error and response...
}
```

## Manager Pattern

The `CargoApplicationManager` wires all use cases together:

```go
func NewCargoApplicationManager(
    cargoRepo CargoRepository,
    trackingRepo TrackingRepository,
    vesselReader VesselReader,
    publisher EventPublisher,
) *CargoApplicationManager {
    return &CargoApplicationManager{
        CreateCargo:        NewCreateCargoUseCase(cargoRepo),
        GetCargo:           NewGetCargoUseCase(cargoRepo),
        ListCargos:         NewListCargosUseCase(cargoRepo),
        ListCargosByVessel: NewListCargosByVesselUseCase(cargoRepo),
        UpdateStatus:       NewUpdateCargoStatusUseCase(cargoRepo, trackingRepo, vesselReader, publisher),
    }
}
```

Benefits of manager pattern:
- Single point of dependency injection
- Easy to compose complex use cases
- Clear visibility into what uses what repositories

## Key Interfaces

### CargoRepository
```go
type CargoRepository interface {
    Create(ctx context.Context, input CreateCargoInput) (*Cargo, error)
    GetByID(ctx context.Context, id uuid.UUID) (*Cargo, error)
    List(ctx context.Context) ([]*Cargo, error)
    ListByVesselID(ctx context.Context, vesselID uuid.UUID) ([]*Cargo, error)
    UpdateStatus(ctx context.Context, id uuid.UUID, status CargoStatus) (*Cargo, error)
}
```

### TrackingRepository
```go
type TrackingRepository interface {
    Create(ctx context.Context, input AddTrackingInput) (*TrackingEntry, error)
}
```

### EventPublisher
```go
type EventPublisher interface {
    PublishStatusChanged(ctx context.Context, event StatusChangedEvent) error
}
```

## Error Handling

Use cases return domain errors directly - HTTP handlers map them to status codes:

- `cargo.ErrNotFound` → HTTP 404
- `cargo.ErrInvalidInput`, `cargo.ErrInvalidStatus` → HTTP 400
- `cargo.ErrInvalidTransition` → HTTP 422 (Unprocessable Entity)
- Other errors → HTTP 500

## Testing

Tests mock repositories at the use case level:

```go
mockCargoRepo := new(MockCargoRepository)
mockCargoRepo.On("GetByID", mock.Anything, id).
    Return(&cargo.Cargo{...}, nil)

mockVesselReader := new(MockVesselReader)
mockVesselReader.On("GetByID", mock.Anything, mock.Anything).
    Return(&vessel.Vessel{CurrentLocation: "Port of Rotterdam"}, nil)

appManager := appcargo.NewCargoApplicationManager(mockCargoRepo, mockTrackingRepo, mockVesselReader, mockPublisher)
useCase := appManager.GetCargo
result, err := useCase.Execute(ctx, id)
```

## Future Enhancements

- Add transaction support for multi-step operations
- Add result caching at use case level
- Add distributed tracing context propagation
- Add use case-level metrics and observability
- Add saga pattern for eventually-consistent operations
