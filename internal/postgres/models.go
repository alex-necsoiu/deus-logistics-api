package postgres

import (
	"time"

	"github.com/google/uuid"
)

// Cargo is the database model for cargo records.
// Internal use only - never exposed in API responses.
type Cargo struct {
	ID          uuid.UUID
	Name        string
	Description string
	Weight      float64
	Status      string
	VesselID    uuid.UUID
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Vessel is the database model for vessel records.
type Vessel struct {
	ID              uuid.UUID
	Name            string
	Capacity        float64
	CurrentLocation string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// TrackingEntry is the database model for tracking entries.
type TrackingEntry struct {
	ID        uuid.UUID
	CargoID   uuid.UUID
	Location  string
	Status    string
	Note      string
	Timestamp time.Time
}

// CargoEvent is the database model for cargo events.
type CargoEvent struct {
	ID        uuid.UUID
	CargoID   uuid.UUID
	OldStatus string
	NewStatus string
	Timestamp time.Time
}
