package tracking

import (
	"context"

	"github.com/google/uuid"
)

// Repository defines the interface for tracking data persistence.
// Implemented by internal/postgres/tracking_repo.go.
// This table is APPEND-ONLY — never update or delete entries.
type Repository interface {
	// Append persists a new tracking entry to the immutable log. APPEND-ONLY.
	// This is the ONLY way to write to tracking — no updates or deletes allowed.
	// Returns error if cargo does not exist or DB operation fails.
	Append(ctx context.Context, input AddTrackingInput) (*TrackingEntry, error)

	// ListByCargoID retrieves all tracking entries for a cargo in chronological order.
	// Returns empty slice if no entries found.
	ListByCargoID(ctx context.Context, cargoID uuid.UUID) ([]*TrackingEntry, error)
}
