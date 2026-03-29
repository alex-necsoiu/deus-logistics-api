package cargo

import (
	"context"

	domaincargo "github.com/alex-necsoiu/deus-logistics-api/internal/domain/cargo"
	"github.com/alex-necsoiu/deus-logistics-api/internal/domain/tracking"
	"github.com/google/uuid"
)

// CargoRepository defines persistence operations for cargo.
type CargoRepository interface {
	Create(ctx context.Context, input domaincargo.CreateCargoInput) (*domaincargo.Cargo, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domaincargo.Cargo, error)
	List(ctx context.Context) ([]*domaincargo.Cargo, error)
	ListByVesselID(ctx context.Context, vesselID uuid.UUID) ([]*domaincargo.Cargo, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status domaincargo.CargoStatus) (*domaincargo.Cargo, error)
}

// TrackingRepository defines persistence operations for tracking events.
// APPEND-ONLY CONSTRAINT: Only Append() is allowed. No updates or deletes.
type TrackingRepository interface {
	// Append writes a new tracking entry to the immutable append-only log.
	// This is the ONLY write operation allowed on tracking entries.
	Append(ctx context.Context, input tracking.AddTrackingInput) (*tracking.TrackingEntry, error)
}

// EventPublisher defines event publishing operations.
type EventPublisher interface {
	PublishStatusChanged(ctx context.Context, event domaincargo.StatusChangedEvent) error
}
