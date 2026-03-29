package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/alex-necsoiu/deus-logistics-api/internal/domain/cargo"
	"github.com/alex-necsoiu/deus-logistics-api/internal/domain/tracking"
)

// CargoService implements cargo.Service with business logic.
type CargoService struct {
	repo      cargo.Repository
	publisher cargo.EventPublisher
	tracker   tracking.Repository
}

// NewCargoService creates a new cargo service instance.
func NewCargoService(
	repo cargo.Repository,
	publisher cargo.EventPublisher,
	tracker tracking.Repository,
) *CargoService {
	return &CargoService{
		repo:      repo,
		publisher: publisher,
		tracker:   tracker,
	}
}

// CreateCargo registers a new cargo shipment assigned to a vessel.
func (s *CargoService) CreateCargo(ctx context.Context, input cargo.CreateCargoInput) (*cargo.Cargo, error) {
	c, err := s.repo.Create(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("createCargo: %w", err)
	}
	zerolog.Ctx(ctx).Info().Str("cargo_id", c.ID.String()).Str("name", c.Name).Msg("cargo created")
	return c, nil
}

// GetCargo retrieves a cargo by ID.
func (s *CargoService) GetCargo(ctx context.Context, id uuid.UUID) (*cargo.Cargo, error) {
	if id == uuid.Nil {
		return nil, cargo.ErrInvalidInput
	}
	c, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("getCargo: %w", err)
	}
	return c, nil
}

// ListCargoes retrieves all cargo records.
func (s *CargoService) ListCargoes(ctx context.Context) ([]*cargo.Cargo, error) {
	cargoes, err := s.repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("listCargoes: %w", err)
	}
	if cargoes == nil {
		cargoes = []*cargo.Cargo{}
	}
	return cargoes, nil
}

// ListCargoByVesselID retrieves all cargo records for a specific vessel.
func (s *CargoService) ListCargoByVesselID(ctx context.Context, vesselID uuid.UUID) ([]*cargo.Cargo, error) {
	if vesselID == uuid.Nil {
		return nil, cargo.ErrInvalidInput
	}
	cargoes, err := s.repo.ListByVesselID(ctx, vesselID)
	if err != nil {
		return nil, fmt.Errorf("listCargoByVesselID: %w", err)
	}
	if cargoes == nil {
		cargoes = []*cargo.Cargo{}
	}
	return cargoes, nil
}

// UpdateCargoStatus transitions cargo to a new status using domain-enforced state machine.
func (s *CargoService) UpdateCargoStatus(ctx context.Context, id uuid.UUID, status cargo.CargoStatus) (*cargo.Cargo, error) {
	// Quick validation: status must be a recognized value before attempting database lookup
	if !status.IsValid() {
		return nil, cargo.ErrInvalidStatus
	}

	// Get current cargo
	currentCargo, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("updateCargoStatus: %w", err)
	}

	// Enforce state machine transitions at domain level
	oldStatus := currentCargo.Status
	if err := currentCargo.UpdateStatus(status); err != nil {
		// Return domain-level error directly (invalid transition)
		return nil, err
	}

	// Persist transitioned cargo to database
	updatedCargo, err := s.repo.UpdateStatus(ctx, id, currentCargo.Status)
	if err != nil {
		return nil, fmt.Errorf("updateCargoStatus: %w", err)
	}

	// Create tracking record for the status change
	if s.tracker != nil {
		trackingInput := tracking.AddTrackingInput{CargoID: id, Location: "Unknown", Status: status.String(), Note: fmt.Sprintf("Status changed from %s to %s", oldStatus.String(), status.String())}
		_, _ = s.tracker.Create(ctx, trackingInput)
	}

	// Fire-and-forget Kafka publish: errors are logged by producer but must not fail the HTTP response.
	// This ensures Kafka outages don't impact the API SLA while maintaining observability.
	if s.publisher != nil {
		event := cargo.StatusChangedEvent{ID: uuid.New().String(), EventType: cargo.EventTypeStatusChanged, CargoID: id.String(), OldStatus: oldStatus, NewStatus: status, Timestamp: updatedCargo.UpdatedAt}
		_ = s.publisher.PublishStatusChanged(ctx, event)
	}

	zerolog.Ctx(ctx).Info().Str("cargo_id", id.String()).Str("old_status", oldStatus.String()).Str("new_status", status.String()).Msg("cargo status updated")
	return updatedCargo, nil
}
