package vessel

import "errors"

// Domain-level errors for vessel operations.
// These represent business logic failures, not infrastructure failures.
// Use errors.Is() to check error types.
var (
	// ErrNotFound is returned when a vessel cannot be found.
	// HTTP Status: 404 Not Found
	ErrNotFound = errors.New("vessel not found")

	// ErrInvalidInput is returned when request input validation fails.
	// Wraps all validation failures (missing fields, invalid formats, etc.)
	// HTTP Status: 400 Bad Request
	ErrInvalidInput = errors.New("invalid input")

	// ErrCapacityExceeded is returned when cargo exceeds vessel capacity.
	ErrCapacityExceeded = errors.New("vessel capacity exceeded")
)
