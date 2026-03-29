package tracking

import (
	"context"

	"github.com/google/uuid"
)

// Service defines the interface for tracking business logic.
// Implemented by internal/service/tracking_service.go.
type Service interface {
	// AddTrackingEntry appends an immutable tracking entry for a cargo.
	// Returns ErrInvalidEntry if validation fails.
	// Returns ErrCargoNotFound if cargo does not exist.
	AddTrackingEntry(ctx context.Context, input AddTrackingInput) (*TrackingEntry, error)

	// GetTrackingHistory retrieves all tracking entries for a cargo in chronological order.
	// Returns empty slice if no entries found.
	// Returns ErrCargoNotFound if cargo does not exist.
	GetTrackingHistory(ctx context.Context, cargoID uuid.UUID) ([]*TrackingEntry, error)
}
