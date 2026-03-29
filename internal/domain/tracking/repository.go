package tracking

import (
	"context"

	"github.com/google/uuid"
)

// Repository defines the interface for tracking data persistence.
// Implemented by internal/postgres/tracking_repo.go.
// This table is APPEND-ONLY — never update or delete entries.
type Repository interface {
	// Create persists a new tracking entry. APPEND-ONLY.
	// Returns error if cargo does not exist or DB operation fails.
	Create(ctx context.Context, input AddTrackingInput) (*TrackingEntry, error)

	// ListByCargoID retrieves all tracking entries for a cargo in chronological order.
	// Returns empty slice if no entries found.
	ListByCargoID(ctx context.Context, cargoID uuid.UUID) ([]*TrackingEntry, error)
}
