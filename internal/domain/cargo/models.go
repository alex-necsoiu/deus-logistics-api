package cargo

import (
	"time"

	"github.com/google/uuid"
)

// CargoStatus represents the current state of a cargo shipment.
type CargoStatus string

const (
	// CargoStatusPending indicates cargo is registered but not yet in transit.
	CargoStatusPending CargoStatus = "pending"
	// CargoStatusInTransit indicates cargo is currently being transported.
	CargoStatusInTransit CargoStatus = "in_transit"
	// CargoStatusDelivered indicates cargo has been delivered to its destination.
	CargoStatusDelivered CargoStatus = "delivered"
)

// IsValid checks if the status is one of the allowed values.
func (s CargoStatus) IsValid() bool {
	switch s {
	case CargoStatusPending, CargoStatusInTransit, CargoStatusDelivered:
		return true
	default:
		return false
	}
}

// String returns the string representation of CargoStatus.
func (s CargoStatus) String() string { return string(s) }

// Cargo represents a shipment of goods assigned to a vessel.
// This is the pure domain model — never expose sqlc-generated types here.
type Cargo struct {
	ID          uuid.UUID
	Name        string
	Description string
	Weight      float64
	Status      CargoStatus
	VesselID    uuid.UUID
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// IsDelivered returns true if the cargo has reached its destination.
func (c *Cargo) IsDelivered() bool { return c.Status == CargoStatusDelivered }

// IsInTransit returns true if the cargo is currently being transported.
func (c *Cargo) IsInTransit() bool { return c.Status == CargoStatusInTransit }

// CreateCargoInput contains validated input for creating a new cargo record.
type CreateCargoInput struct {
	Name        string
	Description string
	Weight      float64
	VesselID    uuid.UUID
}

// Validate checks that all required fields for cargo creation are present and valid.
func (i CreateCargoInput) Validate() error {
	switch {
	case i.Name == "":
		return ErrInvalidInput
	case i.Weight <= 0:
		return ErrInvalidInput
	case i.VesselID == (uuid.UUID{}):
		return ErrInvalidInput
	default:
		return nil
	}
}
