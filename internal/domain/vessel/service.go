package vessel

import (
	"context"

	"github.com/google/uuid"
)

// Service defines the interface for vessel business logic.
// Implemented by internal/service/vessel_service.go.
type Service interface {
	// CreateVessel registers a new vessel.
	// Returns ErrInvalidInput if validation fails.
	CreateVessel(ctx context.Context, input CreateVesselInput) (*Vessel, error)

	// GetVessel retrieves a vessel by ID.
	// Returns ErrNotFound if vessel does not exist.
	GetVessel(ctx context.Context, id uuid.UUID) (*Vessel, error)

	// ListVessels retrieves all vessel records.
	ListVessels(ctx context.Context) ([]*Vessel, error)

	// UpdateVesselLocation updates a vessel's current location.
	// Returns ErrNotFound if vessel does not exist.
	UpdateVesselLocation(ctx context.Context, id uuid.UUID, location string) (*Vessel, error)
}
