package tracking

import "errors"

// Domain-level errors for tracking operations.
var (
	// ErrInvalidEntry is returned when tracking input validation fails.
	ErrInvalidEntry = errors.New("invalid tracking entry")
	// ErrCargoNotFound is returned when cargo does not exist.
	ErrCargoNotFound = errors.New("cargo not found")
)
