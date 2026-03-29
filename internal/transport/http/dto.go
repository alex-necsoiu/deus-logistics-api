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
	Name        string    `json:"name" binding:"required" example:"Premium Electronics" description:"Cargo name"`
	Description string    `json:"description" example:"Laptop shipment" description:"Optional cargo description"`
	Weight      float64   `json:"weight" binding:"required,gt=0" example:"150.5" description:"Cargo weight in kg"`
	VesselID    uuid.UUID `json:"vessel_id" binding:"required" example:"123e4567-e89b-12d3-a456-426614174000" description:"UUID of the vessel carrying this cargo"`
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
	Status string `json:"status" binding:"required" example:"in_transit" description:"Cargo status (pending, in_transit, or delivered)"`
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
	Name            string  `json:"name" binding:"required" example:"MV Ocean Navigator" description:"Vessel name"`
	Capacity        float64 `json:"capacity" binding:"required,gt=0" example:"500.0" description:"Vessel capacity in tons"`
	CurrentLocation string  `json:"current_location" binding:"required" example:"Port of Singapore" description:"Vessel current location"`
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
	CurrentLocation string `json:"current_location" binding:"required" example:"Port of Rotterdam" description:"New vessel location"`
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
	Location string `json:"location" binding:"required" example:"Port of Mumbai" description:"Cargo location"`
	Status   string `json:"status" binding:"required" example:"in_transit" description:"Cargo status (pending, in_transit, or delivered)"`
	Note     string `json:"note" example:"Passed customs inspection" description:"Optional tracking note"`
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
	ID          uuid.UUID `json:"id" example:"123e4567-e89b-12d3-a456-426614174000" description:"Cargo unique identifier"`
	Name        string    `json:"name" example:"Premium Electronics" description:"Cargo name"`
	Description string    `json:"description" example:"Laptop shipment" description:"Cargo description"`
	Weight      float64   `json:"weight" example:"150.5" description:"Cargo weight in kg"`
	Status      string    `json:"status" example:"in_transit" description:"Current cargo status"`
	VesselID    uuid.UUID `json:"vessel_id" example:"123e4567-e89b-12d3-a456-426614174001" description:"UUID of the vessel carrying this cargo"`
	CreatedAt   time.Time `json:"created_at" example:"2024-01-15T10:30:00Z" description:"Cargo creation timestamp"`
	UpdatedAt   time.Time `json:"updated_at" example:"2024-01-15T14:45:00Z" description:"Last cargo update timestamp"`
}

// VesselResponse is the response DTO for vessel.
type VesselResponse struct {
	ID              uuid.UUID `json:"id" example:"123e4567-e89b-12d3-a456-426614174001" description:"Vessel unique identifier"`
	Name            string    `json:"name" example:"MV Ocean Navigator" description:"Vessel name"`
	Capacity        float64   `json:"capacity" example:"500.0" description:"Vessel capacity in tons"`
	CurrentLocation string    `json:"current_location" example:"Port of Singapore" description:"Vessel current location"`
	CreatedAt       time.Time `json:"created_at" example:"2024-01-01T08:00:00Z" description:"Vessel creation timestamp"`
	UpdatedAt       time.Time `json:"updated_at" example:"2024-01-15T14:45:00Z" description:"Last vessel update timestamp"`
}

// TrackingEntryResponse is the response DTO for a tracking entry.
type TrackingEntryResponse struct {
	ID        uuid.UUID `json:"id" example:"123e4567-e89b-12d3-a456-426614174002" description:"Tracking entry unique identifier"`
	CargoID   uuid.UUID `json:"cargo_id" example:"123e4567-e89b-12d3-a456-426614174000" description:"Associated cargo UUID"`
	Location  string    `json:"location" example:"Port of Mumbai" description:"Cargo location at this timestamp"`
	Status    string    `json:"status" example:"in_transit" description:"Cargo status at this timestamp"`
	Note      string    `json:"note" example:"Passed customs inspection" description:"Optional tracking note"`
	Timestamp time.Time `json:"timestamp" example:"2024-01-15T12:30:00Z" description:"When this tracking entry was recorded"`
}

// CargoEventResponse is the response DTO for a cargo event.
type CargoEventResponse struct {
	ID        uuid.UUID `json:"id" example:"123e4567-e89b-12d3-a456-426614174003" description:"Event unique identifier"`
	CargoID   uuid.UUID `json:"cargo_id" example:"123e4567-e89b-12d3-a456-426614174000" description:"Associated cargo UUID"`
	OldStatus string    `json:"old_status" example:"pending" description:"Previous cargo status"`
	NewStatus string    `json:"new_status" example:"in_transit" description:"New cargo status"`
	Timestamp time.Time `json:"timestamp" example:"2024-01-15T11:00:00Z" description:"When this status change occurred"`
}
