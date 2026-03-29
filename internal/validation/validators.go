package validation

import (
	"fmt"
	"strings"
)

// ValidateRequiredString checks if a string field is non-empty after trimming whitespace.
// Returns an error with standardized message if validation fails.
func ValidateRequiredString(fieldName, value string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("%s cannot be empty", fieldName)
	}
	return nil
}

// ValidatePositiveFloat checks if a float64 value is greater than zero.
// Returns an error with standardized message if validation fails.
func ValidatePositiveFloat(fieldName string, value float64) error {
	if value <= 0 {
		return fmt.Errorf("%s must be greater than 0", fieldName)
	}
	return nil
}

// ValidateFloatRange checks if a float64 value is within [min, max] inclusive.
// Returns an error with standardized message if validation fails.
func ValidateFloatRange(fieldName string, value, min, max float64) error {
	if value < min || value > max {
		return fmt.Errorf("%s must be between %.0f and %.0f", fieldName, min, max)
	}
	return nil
}

// ValidateMaxFloat checks if a float64 value does not exceed a maximum.
// Returns an error with standardized message if validation fails.
func ValidateMaxFloat(fieldName string, value, max float64) error {
	if value > max {
		return fmt.Errorf("%s cannot exceed %.0f", fieldName, max)
	}
	return nil
}

// ValidateCargoStatus checks if a status string is valid for cargo operations.
// Valid statuses: pending, in_transit, delivered
func ValidateCargoStatus(status string) error {
	validStatuses := map[string]bool{
		"pending":    true,
		"in_transit": true,
		"delivered":  true,
	}
	if !validStatuses[status] {
		return fmt.Errorf("invalid status: %s, must be one of: pending, in_transit, delivered", status)
	}
	return nil
}

// ValidateVesselCapacity checks if vessel capacity is within acceptable range.
// Must be positive and not exceed fleet maximum (1000 tons).
func ValidateVesselCapacity(capacity float64) error {
	if capacity <= 0 {
		return fmt.Errorf("capacity must be greater than 0")
	}
	if capacity > 1000 {
		return fmt.Errorf("capacity cannot exceed 1000 tons (fleet maximum)")
	}
	return nil
}

// ValidationErrors collects multiple validation errors for batch reporting.
type ValidationErrors struct {
	errors []string
}

// Add appends an error message to the collection.
func (ve *ValidationErrors) Add(err string) {
	if err != "" {
		ve.errors = append(ve.errors, err)
	}
}

// AddError appends an error to the collection if it is not nil.
func (ve *ValidationErrors) AddError(err error) {
	if err != nil {
		ve.errors = append(ve.errors, err.Error())
	}
}

// Error returns a formatted error message combining all collected errors.
// Returns empty string if no errors were collected.
func (ve *ValidationErrors) Error() string {
	if len(ve.errors) == 0 {
		return ""
	}
	if len(ve.errors) == 1 {
		return ve.errors[0]
	}
	// Multiple errors: format as bullet list
	return fmt.Sprintf("validation failed with %d errors: %s", len(ve.errors), strings.Join(ve.errors, "; "))
}

// Valid returns true if no errors were collected.
func (ve *ValidationErrors) Valid() bool {
	return len(ve.errors) == 0
}

// Count returns the number of validation errors collected.
func (ve *ValidationErrors) Count() int {
	return len(ve.errors)
}
