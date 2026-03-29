package tracking

import "errors"

// Domain-level errors for tracking operations.
// These represent business logic failures, not infrastructure failures.
// Use errors.Is() to check error types.
var (
	// ErrInvalidEntry is returned when tracking input validation fails.
	// Wraps all validation failures (missing cargo, empty location, etc.)
	// HTTP Status: 400 Bad Request
	ErrInvalidEntry = errors.New("invalid tracking entry")

	// ErrCargoNotFound is returned when cargo does not exist.
	// Typically means the cargo_id reference is invalid.
	// HTTP Status: 404 Not Found
	ErrCargoNotFound = errors.New("cargo not found")
)
