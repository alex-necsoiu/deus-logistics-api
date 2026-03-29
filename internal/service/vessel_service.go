package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/alex-necsoiu/deus-logistics-api/internal/domain/vessel"
)

// VesselService implements vessel.Service.
type VesselService struct {
	repo vessel.Repository
}

// NewVesselService creates a new vessel service with the given repository.
//
// Inputs:
//
//	repo - vessel repository implementation (must not be nil)
//
// Returns:
//
//	*VesselService with initialized repository
func NewVesselService(repo vessel.Repository) *VesselService {
	return &VesselService{repo: repo}
}

// CreateVessel registers a new vessel in the system.
//
// Inputs:
//
//	ctx   - request context for cancellation and tracing
//	input - vessel details (Name required, Capacity must be > 0)
//
// Returns:
//
//	*Vessel with generated UUID on success
//	ErrInvalidInput if Name is empty or Capacity <= 0
//
// Side effects:
//   - DB write to vessels table
//   - Logs info message with vessel_id and name
func (s *VesselService) CreateVessel(ctx context.Context, input vessel.CreateVesselInput) (*vessel.Vessel, error) {
	v, err := s.repo.Create(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("createVessel: %w", err)
	}
	zerolog.Ctx(ctx).Info().Str("vessel_id", v.ID.String()).Str("name", v.Name).Msg("vessel created")
	return v, nil
}

// GetVessel retrieves a vessel by ID.
//
// Inputs:
//
//	ctx - request context for cancellation and tracing
//	id  - UUID of the vessel (must not be nil)
//
// Returns:
//
//	*Vessel on success
//	ErrInvalidInput if id is nil
//	ErrNotFound if vessel does not exist
//
// Side effects:
//   - DB read from vessels table
func (s *VesselService) GetVessel(ctx context.Context, id uuid.UUID) (*vessel.Vessel, error) {
	if id == uuid.Nil {
		return nil, vessel.ErrInvalidInput
	}
	v, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("getVessel: %w", err)
	}
	return v, nil
}

// ListVessels retrieves all vessels in the system.
//
// Inputs:
//
//	ctx - request context for cancellation and tracing
//
// Returns:
//
//	[]*Vessel sorted by creation timestamp on success
//	Empty slice if no vessels exist
//
// Side effects:
//   - DB read from vessels table
func (s *VesselService) ListVessels(ctx context.Context) ([]*vessel.Vessel, error) {
	vessels, err := s.repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("listVessels: %w", err)
	}
	if vessels == nil {
		vessels = []*vessel.Vessel{}
	}
	return vessels, nil
}

// UpdateVesselLocation updates the current location of a vessel.
//
// Inputs:
//
//	ctx      - request context for cancellation and tracing
//	id       - UUID of the vessel (must not be nil)
//	location - new location string (must not be empty)
//
// Returns:
//
//	*Vessel with updated location on success
//	ErrInvalidInput if id is nil or location is empty
//	ErrNotFound if vessel does not exist
//
// Side effects:
//   - DB update to vessels table (current_location column)
//   - Logs info message with vessel_id and location
func (s *VesselService) UpdateVesselLocation(ctx context.Context, id uuid.UUID, location string) (*vessel.Vessel, error) {
	if id == uuid.Nil || location == "" {
		return nil, vessel.ErrInvalidInput
	}
	v, err := s.repo.UpdateLocation(ctx, id, location)
	if err != nil {
		return nil, fmt.Errorf("updateVesselLocation: %w", err)
	}
	zerolog.Ctx(ctx).Info().Str("vessel_id", id.String()).Str("location", location).Msg("vessel location updated")
	return v, nil
}

// UpdateVesselCapacity updates the cargo capacity of a vessel.
//
// Inputs:
//
//	ctx      - request context for cancellation and tracing
//	id       - UUID of the vessel (must not be nil)
//	capacity - new capacity in units (must be > 0)
//
// Returns:
//
//	*Vessel with updated capacity on success
//	ErrInvalidInput if id is nil or capacity <= 0
//	ErrNotFound if vessel does not exist
//
// Side effects:
//   - DB update to vessels table (capacity column)
//   - Logs info message with vessel_id and capacity
func (s *VesselService) UpdateVesselCapacity(ctx context.Context, id uuid.UUID, capacity float64) (*vessel.Vessel, error) {
	if id == uuid.Nil || capacity <= 0 {
		return nil, vessel.ErrInvalidInput
	}
	v, err := s.repo.UpdateCapacity(ctx, id, capacity)
	if err != nil {
		return nil, fmt.Errorf("updateVesselCapacity: %w", err)
	}
	zerolog.Ctx(ctx).Info().Str("vessel_id", id.String()).Float64("capacity", capacity).Msg("vessel capacity updated")
	return v, nil
}
