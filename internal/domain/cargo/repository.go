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

	// ListByVesselID retrieves all cargo records for a specific vessel.
	ListByVesselID(ctx context.Context, vesselID uuid.UUID) ([]*Cargo, error)

	// UpdateStatus updates the cargo status field.
	// Returns ErrNotFound if cargo does not exist.
	UpdateStatus(ctx context.Context, id uuid.UUID, status CargoStatus) (*Cargo, error)
}
