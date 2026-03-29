// Package errors provides standardized error handling and mapping for the DEUS Logistics API.
package errors

import (
	"errors"
	"net/http"

	"github.com/alex-necsoiu/deus-logistics-api/internal/domain/cargo"
	"github.com/alex-necsoiu/deus-logistics-api/internal/domain/tracking"
	"github.com/alex-necsoiu/deus-logistics-api/internal/domain/vessel"
)

// Domain-Level Sentinel Errors
// These represent business logic failures that should be handled gracefully by clients.
// They are never wrapped - sentinel values only.
var (
	// ErrNotFound indicates a requested resource does not exist.
	// HTTP Status: 404 Not Found
	ErrNotFound = errors.New("resource not found")

	// ErrInvalidTransition indicates a state transition is not allowed by business rules.
	// HTTP Status: 422 Unprocessable Entity
	ErrInvalidTransition = errors.New("invalid state transition")

	// ErrInvalidInput indicates request input validation failed.
	// HTTP Status: 400 Bad Request
	ErrInvalidInput = errors.New("invalid input")

	// ErrUnknown indicates an unexpected error occurred outside business logic.
	// HTTP Status: 500 Internal Server Error
	ErrUnknown = errors.New("internal server error")
)

// ErrorCode represents a machine-readable error classification for API clients.
type ErrorCode string

const (
	// CodeNotFound: Resource does not exist
	CodeNotFound ErrorCode = "NOT_FOUND"

	// CodeInvalidInput: Request validation failed
	CodeInvalidInput ErrorCode = "INVALID_INPUT"

	// CodeInvalidTransition: Business logic rejected operation
	CodeInvalidTransition ErrorCode = "INVALID_TRANSITION"

	// CodeInternalError: Unexpected server error
	CodeInternalError ErrorCode = "INTERNAL_ERROR"
)

// MapErrorToHTTPStatus maps domain errors to appropriate HTTP status codes.
// Checks both central error package sentinels and domain-specific errors from all packages.
func MapErrorToHTTPStatus(err error) int {
	if err == nil {
		return http.StatusOK
	}

	// Check for "not found" errors across all domains
	if errors.Is(err, ErrNotFound) || errors.Is(err, cargo.ErrNotFound) ||
		errors.Is(err, vessel.ErrNotFound) || errors.Is(err, tracking.ErrCargoNotFound) {
		return http.StatusNotFound
	}

	// Check for invalid transition errors
	if errors.Is(err, ErrInvalidTransition) || errors.Is(err, cargo.ErrInvalidTransition) {
		return http.StatusUnprocessableEntity
	}

	// Check for invalid input errors (includes all input validation failures)
	if errors.Is(err, ErrInvalidInput) || errors.Is(err, cargo.ErrInvalidInput) ||
		errors.Is(err, cargo.ErrInvalidStatus) || errors.Is(err, vessel.ErrInvalidInput) ||
		errors.Is(err, tracking.ErrInvalidEntry) {
		return http.StatusBadRequest
	}

	// Default to internal server error for unhandled cases
	return http.StatusInternalServerError
}

// MapErrorToErrorCode maps domain errors to API error codes.
// Checks both central error package sentinels and domain-specific errors from all packages.
func MapErrorToErrorCode(err error) ErrorCode {
	if err == nil {
		return ""
	}

	// Check for "not found" errors across all domains
	if errors.Is(err, ErrNotFound) || errors.Is(err, cargo.ErrNotFound) ||
		errors.Is(err, vessel.ErrNotFound) || errors.Is(err, tracking.ErrCargoNotFound) {
		return CodeNotFound
	}

	// Check for invalid transition errors
	if errors.Is(err, ErrInvalidTransition) || errors.Is(err, cargo.ErrInvalidTransition) {
		return CodeInvalidTransition
	}

	// Check for invalid input errors (includes all input validation failures)
	if errors.Is(err, ErrInvalidInput) || errors.Is(err, cargo.ErrInvalidInput) ||
		errors.Is(err, cargo.ErrInvalidStatus) || errors.Is(err, vessel.ErrInvalidInput) ||
		errors.Is(err, tracking.ErrInvalidEntry) {
		return CodeInvalidInput
	}

	return CodeInternalError
}

// MapErrorToErrorMessage returns user-safe error message.
// Never exposes database details, SQL, or stack traces.
// Checks both central error package sentinels and domain-specific errors from all packages.
func MapErrorToErrorMessage(err error) string {
	if err == nil {
		return ""
	}

	// Check for "not found" errors across all domains
	if errors.Is(err, ErrNotFound) || errors.Is(err, cargo.ErrNotFound) ||
		errors.Is(err, vessel.ErrNotFound) || errors.Is(err, tracking.ErrCargoNotFound) {
		return "The requested resource was not found."
	}

	// Check for invalid transition errors
	if errors.Is(err, ErrInvalidTransition) || errors.Is(err, cargo.ErrInvalidTransition) {
		return "The requested operation violates business rules or state constraints."
	}

	// Check for invalid input errors (includes all input validation failures)
	if errors.Is(err, ErrInvalidInput) || errors.Is(err, cargo.ErrInvalidInput) ||
		errors.Is(err, cargo.ErrInvalidStatus) || errors.Is(err, vessel.ErrInvalidInput) ||
		errors.Is(err, tracking.ErrInvalidEntry) {
		return "The request input is invalid. Please check the request and try again."
	}

	return "An unexpected error occurred. Please try again later."
}
