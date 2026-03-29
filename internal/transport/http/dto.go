package http

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// --- Request DTOs ---

// CreateCargoRequest is the request body for creating a cargo.
type CreateCargoRequest struct {
	Name        string    `json:"name" binding:"required"`
	Description string    `json:"description"`
	Weight      float64   `json:"weight" binding:"required,gt=0"`
	VesselID    uuid.UUID `json:"vessel_id" binding:"required"`
}

// Validate performs business validation on the cargo creation request.
// Returns nil if all fields are valid, otherwise returns a descriptive error.
func (r *CreateCargoRequest) Validate() error {
	if strings.TrimSpace(r.Name) == "" {
		return errors.New("name cannot be empty")
	}
	if r.Weight <= 0 {
		return errors.New("weight must be greater than 0")
	}
	return nil
}

// UpdateCargoStatusRequest is the request body for updating cargo status.
type UpdateCargoStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

// Validate performs business validation on the status update.
// Returns nil if status is valid, otherwise returns a descriptive error.
func (r *UpdateCargoStatusRequest) Validate() error {
	validStatuses := map[string]bool{
		"pending":    true,
		"in_transit": true,
		"delivered":  true,
	}
	if !validStatuses[r.Status] {
		return fmt.Errorf("invalid status: %s, must be one of: pending, in_transit, delivered", r.Status)
	}
	return nil
}

// CreateVesselRequest is the request body for creating a vessel.
type CreateVesselRequest struct {
	Name            string  `json:"name" binding:"required"`
	Capacity        float64 `json:"capacity" binding:"required,gt=0"`
	CurrentLocation string  `json:"current_location" binding:"required"`
}

// Validate performs business validation on the vessel creation request.
// Returns nil if all fields are valid, otherwise returns a descriptive error.
func (r *CreateVesselRequest) Validate() error {
	if strings.TrimSpace(r.Name) == "" {
		return errors.New("name cannot be empty")
	}
	if r.Capacity <= 0 {
		return errors.New("capacity must be greater than 0")
	}
	if r.Capacity > 1000 {
		return errors.New("capacity cannot exceed 1000 tons (fleet maximum)")
	}
	if strings.TrimSpace(r.CurrentLocation) == "" {
		return errors.New("current location cannot be empty")
	}
	return nil
}

// UpdateVesselLocationRequest is the request body for updating vessel location.
type UpdateVesselLocationRequest struct {
	CurrentLocation string `json:"current_location" binding:"required"`
}

// Validate performs business validation on the location update.
// Returns nil if location is valid, otherwise returns a descriptive error.
func (r *UpdateVesselLocationRequest) Validate() error {
	if strings.TrimSpace(r.CurrentLocation) == "" {
		return errors.New("current location cannot be empty")
	}
	return nil
}

// AddTrackingRequest is the request body for adding a tracking entry.
type AddTrackingRequest struct {
	Location string `json:"location" binding:"required"`
	Status   string `json:"status" binding:"required"`
	Note     string `json:"note"`
}

// Validate performs business validation on the tracking entry request.
// Returns nil if all fields are valid, otherwise returns a descriptive error.
func (r *AddTrackingRequest) Validate() error {
	if strings.TrimSpace(r.Location) == "" {
		return errors.New("location cannot be empty")
	}

	validStatuses := map[string]bool{
		"pending":    true,
		"in_transit": true,
		"delivered":  true,
	}
	if !validStatuses[r.Status] {
		return fmt.Errorf("invalid status: %s, must be one of: pending, in_transit, delivered", r.Status)
	}
	return nil
}

// --- Response DTOs ---

// CargoResponse is the response DTO for cargo.
type CargoResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Weight      float64   `json:"weight"`
	Status      string    `json:"status"`
	VesselID    uuid.UUID `json:"vessel_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// VesselResponse is the response DTO for vessel.
type VesselResponse struct {
	ID              uuid.UUID `json:"id"`
	Name            string    `json:"name"`
	Capacity        float64   `json:"capacity"`
	CurrentLocation string    `json:"current_location"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// TrackingEntryResponse is the response DTO for a tracking entry.
type TrackingEntryResponse struct {
	ID        uuid.UUID `json:"id"`
	CargoID   uuid.UUID `json:"cargo_id"`
	Location  string    `json:"location"`
	Status    string    `json:"status"`
	Note      string    `json:"note"`
	Timestamp time.Time `json:"timestamp"`
}

// CargoEventResponse is the response DTO for a cargo event.
type CargoEventResponse struct {
	ID        uuid.UUID `json:"id"`
	CargoID   uuid.UUID `json:"cargo_id"`
	OldStatus string    `json:"old_status"`
	NewStatus string    `json:"new_status"`
	Timestamp time.Time `json:"timestamp"`
}
