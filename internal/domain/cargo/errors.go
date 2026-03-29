package cargo

import (
	"errors"
)

// Domain-level errors for cargo operations.
// These represent business logic failures, not infrastructure failures.
// Sentinel errors are re-exported from central errors package for backward compatibility.
//
// Pattern: Use errors.Is(err, ErrNotFound) to check error type
// Never use: if err.Error() == "..."
var (
	// ErrNotFound is returned when a cargo cannot be found.
	// Alias for central errors package
	ErrNotFound = errors.New("cargo not found")

	// ErrInvalidStatus is returned when an invalid status value is provided.
	// Cargo status must be one of: pending, in_transit, delivered
	ErrInvalidStatus = errors.New("invalid cargo status")

	// ErrInvalidInput is returned when request input validation fails.
	// Wraps all validation failures (missing fields, invalid formats, etc.)
	ErrInvalidInput = errors.New("invalid input")

	// ErrInvalidTransition is returned when a cargo status transition is not allowed.
	// State machine enforces: pending→in_transit→delivered (no backward transitions)
	ErrInvalidTransition = errors.New("invalid status transition")
)
