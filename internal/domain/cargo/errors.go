package cargo

import "errors"

// Domain-level errors for cargo operations.
// These represent business logic failures, not infrastructure failures.
var (
	// ErrNotFound is returned when a cargo cannot be found.
	ErrNotFound = errors.New("cargo not found")
	// ErrInvalidStatus is returned when an invalid status value is provided.
	ErrInvalidStatus = errors.New("invalid cargo status")
	// ErrInvalidInput is returned when request input validation fails.
	ErrInvalidInput = errors.New("invalid input")
	// ErrInvalidTransition is returned when a cargo status transition is not allowed.
	ErrInvalidTransition = errors.New("invalid status transition")
)
