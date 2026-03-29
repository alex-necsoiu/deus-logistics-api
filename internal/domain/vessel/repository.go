package vessel

import (
	"context"

	"github.com/google/uuid"
)

// Repository defines the interface for vessel data persistence.
// Implemented by internal/postgres/vessel_repo.go.
type Repository interface {
	// Create persists a new vessel record and returns the created entity.
	Create(ctx context.Context, input CreateVesselInput) (*Vessel, error)

	// GetByID retrieves a vessel by its unique ID.
	// Returns ErrNotFound if vessel does not exist.
	GetByID(ctx context.Context, id uuid.UUID) (*Vessel, error)

	// List retrieves all vessel records. Returns empty slice if none found.
	List(ctx context.Context) ([]*Vessel, error)

	// UpdateLocation updates the vessel's current location.
	// Returns ErrNotFound if vessel does not exist.
	UpdateLocation(ctx context.Context, id uuid.UUID, location string) (*Vessel, error)

	// UpdateCapacity updates the vessel's cargo capacity.
	// Returns ErrNotFound if vessel does not exist.
	UpdateCapacity(ctx context.Context, id uuid.UUID, capacity float64) (*Vessel, error)
}
