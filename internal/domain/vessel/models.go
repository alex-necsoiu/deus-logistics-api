package vessel

import (
	"time"

	"github.com/google/uuid"
)

// Vessel represents a ship that carries cargo.
// This is the pure domain model — never expose sqlc-generated types here.
type Vessel struct {
	ID               uuid.UUID
	Name             string
	Capacity         float64
	CurrentLocation  string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// CreateVesselInput contains validated input for creating a new vessel record.
type CreateVesselInput struct {
	Name             string
	Capacity         float64
	CurrentLocation  string
}

// Validate checks that all required fields for vessel creation are present and valid.
func (i CreateVesselInput) Validate() error {
	switch {
	case i.Name == "":
		return ErrInvalidInput
	case i.Capacity <= 0:
		return ErrInvalidInput
	case i.CurrentLocation == "":
		return ErrInvalidInput
	default:
		return nil
	}
}

// UpdateVesselLocationInput contains validated input for updating vessel location.
type UpdateVesselLocationInput struct {
	CurrentLocation string
}

// CanCargoFit checks if a cargo of given weight can fit on this vessel.
func (v *Vessel) CanCargoFit(weight float64) bool {
	return weight > 0 && weight <= v.Capacity
}
