package vessel

import "errors"

// Domain-level errors for vessel operations.
var (
	// ErrNotFound is returned when a vessel cannot be found.
	ErrNotFound = errors.New("vessel not found")
	// ErrInvalidInput is returned when request input validation fails.
	ErrInvalidInput = errors.New("invalid input")
	// ErrCapacityExceeded is returned when cargo exceeds vessel capacity.
	ErrCapacityExceeded = errors.New("vessel capacity exceeded")
)
