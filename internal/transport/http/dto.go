package http

import (
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/alex-necsoiu/deus-logistics-api/internal/validation"
)

// --- Request DTOs ---

// CreateCargoRequest is the request body for creating a cargo.
// All fields are required except Description.
// Constraints:
//   - Name: non-empty string
//   - Weight: positive float64 (greater than 0)
//   - VesselID: valid UUID (required)
//   - Description: optional, can be empty
type CreateCargoRequest struct {
	Name        string    `json:"name" binding:"required"`
	Description string    `json:"description"`
	Weight      float64   `json:"weight" binding:"required,gt=0"`
	VesselID    uuid.UUID `json:"vessel_id" binding:"required"`
}

// Validate performs strict business validation on the cargo creation request.
// Validates request constraints BEFORE use case execution.
// Returns nil if all fields are valid, otherwise returns a descriptive error.
func (r *CreateCargoRequest) Validate() error {
	if err := validation.ValidateRequiredString("name", r.Name); err != nil {
		return err
	}
	if err := validation.ValidatePositiveFloat("weight", r.Weight); err != nil {
		return err
	}
	if r.VesselID == uuid.Nil {
		return errors.New("vessel_id is required and must be a valid UUID")
	}
	return nil
}

// UpdateCargoStatusRequest is the request body for updating cargo status.
// Status must be one of: pending, in_transit, delivered
type UpdateCargoStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

// Validate performs strict business validation on the status update.
// Validates that status is a valid cargo state BEFORE use case execution.
// Returns nil if status is valid, otherwise returns a descriptive error.
func (r *UpdateCargoStatusRequest) Validate() error {
	return validation.ValidateCargoStatus(r.Status)
}

// CreateVesselRequest is the request body for creating a vessel.
// All fields are required.
// Constraints:
//   - Name: non-empty string
//   - Capacity: positive float64, max 1000 tons (fleet maximum)
//   - CurrentLocation: non-empty string
type CreateVesselRequest struct {
	Name            string  `json:"name" binding:"required"`
	Capacity        float64 `json:"capacity" binding:"required,gt=0"`
	CurrentLocation string  `json:"current_location" binding:"required"`
}

// Validate performs strict business validation on the vessel creation request.
// Validates all constraints BEFORE use case execution.
// Returns nil if all fields are valid, otherwise returns a descriptive error.
func (r *CreateVesselRequest) Validate() error {
	if err := validation.ValidateRequiredString("name", r.Name); err != nil {
		return err
	}
	if err := validation.ValidateVesselCapacity(r.Capacity); err != nil {
		return err
	}
	if err := validation.ValidateRequiredString("current_location", r.CurrentLocation); err != nil {
		return err
	}
	return nil
}

// UpdateVesselLocationRequest is the request body for updating vessel location.
// Location field is required and cannot be empty.
type UpdateVesselLocationRequest struct {
	CurrentLocation string `json:"current_location" binding:"required"`
}

// Validate performs strict business validation on the location update.
// Validates that location is non-empty BEFORE use case execution.
// Returns nil if location is valid, otherwise returns a descriptive error.
func (r *UpdateVesselLocationRequest) Validate() error {
	return validation.ValidateRequiredString("current_location", r.CurrentLocation)
}

// AddTrackingRequest is the request body for adding a tracking entry.
// Location and Status are required.
// Constraints:
//   - Location: non-empty string
//   - Status: one of pending, in_transit, delivered
//   - Note: optional
type AddTrackingRequest struct {
	Location string `json:"location" binding:"required"`
	Status   string `json:"status" binding:"required"`
	Note     string `json:"note"`
}

// Validate performs strict business validation on the tracking entry request.
// Validates all constraints BEFORE use case execution.
// Returns nil if all fields are valid, otherwise returns a descriptive error.
func (r *AddTrackingRequest) Validate() error {
	if err := validation.ValidateRequiredString("location", r.Location); err != nil {
		return err
	}
	if err := validation.ValidateCargoStatus(r.Status); err != nil {
		return err
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
