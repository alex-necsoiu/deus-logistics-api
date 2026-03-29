package cargo

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/rs/zerolog"

	domaincargo "github.com/alex-necsoiu/deus-logistics-api/internal/domain/cargo"
	"github.com/alex-necsoiu/deus-logistics-api/internal/domain/tracking"
)

// UpdateCargoStatusUseCase transitions cargo to a new status using domain-enforced state machine.
type UpdateCargoStatusUseCase struct {
	cargoRepo    CargoRepository
	trackingRepo TrackingRepository
	publisher    EventPublisher
}

// NewUpdateCargoStatusUseCase creates a new use case with injected dependencies.
func NewUpdateCargoStatusUseCase(
	cargoRepo CargoRepository,
	trackingRepo TrackingRepository,
	publisher EventPublisher,
) *UpdateCargoStatusUseCase {
	return &UpdateCargoStatusUseCase{
		cargoRepo:    cargoRepo,
		trackingRepo: trackingRepo,
		publisher:    publisher,
	}
}

// Execute transitions cargo to a new status.
func (uc *UpdateCargoStatusUseCase) Execute(ctx context.Context, id uuid.UUID, newStatus domaincargo.CargoStatus) (*domaincargo.Cargo, error) {
	// Quick validation: status must be a recognized value
	if !newStatus.IsValid() {
		zerolog.Ctx(ctx).Warn().
			Str("cargo_id", id.String()).
			Str("status", newStatus.String()).
			Msg("invalid cargo status provided")
		return nil, domaincargo.ErrInvalidStatus
	}

	// Retrieve current cargo state
	currentCargo, err := uc.cargoRepo.GetByID(ctx, id)
	if err != nil {
		zerolog.Ctx(ctx).Error().
			Err(err).
			Str("cargo_id", id.String()).
			Msg("failed to retrieve cargo for status update")
		return nil, fmt.Errorf("update_cargo_status: %w", err)
	}

	// Domain-level state machine enforcement: UpdateStatus() encodes all business rules
	oldStatus := currentCargo.Status
	if err := currentCargo.UpdateStatus(newStatus); err != nil {
		// Log invalid state transitions
		zerolog.Ctx(ctx).Warn().
			Str("cargo_id", id.String()).
			Str("current_status", oldStatus.String()).
			Str("requested_status", newStatus.String()).
			Err(err).
			Msg("invalid cargo status transition")
		// Return domain error directly (ErrInvalidTransition or ErrInvalidStatus)
		return nil, err
	}

	// Persist the state transition to database
	updatedCargo, err := uc.cargoRepo.UpdateStatus(ctx, id, newStatus)
	if err != nil {
		zerolog.Ctx(ctx).Error().
			Err(err).
			Str("cargo_id", id.String()).
			Str("old_status", oldStatus.String()).
			Str("new_status", newStatus.String()).
			Msg("failed to persist cargo status update")
		return nil, fmt.Errorf("update_cargo_status: %w", err)
	}

	// Orchestration: append tracking event (not domain concern)
	if uc.trackingRepo != nil {
		trackingInput := tracking.AddTrackingInput{
			CargoID:  id,
			Location: "Unknown",
			Status:   newStatus.String(),
			Note:     fmt.Sprintf("Status changed from %s to %s", oldStatus.String(), newStatus.String()),
		}
		if _, err := uc.trackingRepo.Append(ctx, trackingInput); err != nil {
			// Log tracking error but don't fail the status update
			zerolog.Ctx(ctx).Warn().
				Err(err).
				Str("cargo_id", id.String()).
				Str("status", newStatus.String()).
				Msg("failed to append tracking record after status update")
		}
	}

	// Fire-and-forget event publishing (not domain concern).
	// Errors logged by publisher but don't fail the operation to maintain SLA.
	if uc.publisher != nil {
		event := domaincargo.StatusChangedEvent{
			ID:        uuid.New().String(),
			EventType: domaincargo.EventTypeStatusChanged,
			CargoID:   id.String(),
			OldStatus: oldStatus,
			NewStatus: newStatus,
			Timestamp: updatedCargo.UpdatedAt,
		}
		if err := uc.publisher.PublishStatusChanged(ctx, event); err != nil {
			// Log publish failure but continue
			zerolog.Ctx(ctx).Warn().
				Err(err).
				Str("cargo_id", id.String()).
				Str("event_id", event.ID).
				Str("new_status", newStatus.String()).
				Msg("failed to publish cargo status changed event to kafka")
		} else {
			zerolog.Ctx(ctx).Debug().
				Str("cargo_id", id.String()).
				Str("event_id", event.ID).
				Str("new_status", newStatus.String()).
				Msg("cargo status changed event published to kafka")
		}
	}

	// Log successful transition as business event
	zerolog.Ctx(ctx).Info().
		Str("cargo_id", id.String()).
		Str("old_status", oldStatus.String()).
		Str("new_status", newStatus.String()).
		Msg("business_event:cargo_status_updated")

	return updatedCargo, nil
}
