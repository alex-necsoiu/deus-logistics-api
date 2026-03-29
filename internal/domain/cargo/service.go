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
	CreateCargo(ctx context.Context, input CreateCargoInput) (*Cargo, error)

	// GetCargo retrieves a cargo by ID.
	// Returns ErrNotFound if cargo does not exist.
	GetCargo(ctx context.Context, id uuid.UUID) (*Cargo, error)

	// ListCargoes retrieves all cargo records.
	ListCargoes(ctx context.Context) ([]*Cargo, error)

	// ListCargoByVesselID retrieves all cargo records for a specific vessel.
	ListCargoByVesselID(ctx context.Context, vesselID uuid.UUID) ([]*Cargo, error)

	// UpdateCargoStatus transitions cargo to a new status.
	// Appends an immutable tracking entry on every status change.
	// Emits a Kafka cargo.status_changed event after successful DB write.
	// Returns ErrNotFound if cargo does not exist.
	// Returns ErrInvalidStatus if the provided status is not a valid CargoStatus.
	UpdateCargoStatus(ctx context.Context, id uuid.UUID, status CargoStatus) (*Cargo, error)
}
