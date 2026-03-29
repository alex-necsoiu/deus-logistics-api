package tracking

import (
	"time"

	"github.com/google/uuid"
)

// TrackingEntry represents a single movement record for cargo.
// This table is APPEND-ONLY — entries are never updated or deleted.
type TrackingEntry struct {
	ID        uuid.UUID
	CargoID   uuid.UUID
	Location  string
	Status    string
	Note      string
	Timestamp time.Time
}

// AddTrackingInput contains validated input for creating a tracking entry.
type AddTrackingInput struct {
	CargoID  uuid.UUID
	Location string
	Status   string
	Note     string
}

// Validate checks that all required fields for a tracking entry are present and valid.
func (i AddTrackingInput) Validate() error {
	switch {
	case i.CargoID == (uuid.UUID{}):
		return ErrInvalidEntry
	case i.Location == "":
		return ErrInvalidEntry
	case i.Status == "":
		return ErrInvalidEntry
	default:
		return nil
	}
}
