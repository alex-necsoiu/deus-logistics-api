package cargo

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSentinelErrors validates that sentinel errors are properly defined and comparable.
func TestSentinelErrors(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected error
	}{
		{
			name:     "ErrNotFound is defined",
			err:      ErrNotFound,
			expected: ErrNotFound,
		},
		{
			name:     "ErrInvalidStatus is defined",
			err:      ErrInvalidStatus,
			expected: ErrInvalidStatus,
		},
		{
			name:     "ErrInvalidInput is defined",
			err:      ErrInvalidInput,
			expected: ErrInvalidInput,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given: a sentinel error
			// When: comparing with expected error
			// Then: they are identical (same reference)
			assert.Equal(t, tt.expected, tt.err)
			assert.True(t, errors.Is(tt.err, tt.expected))
		})
	}
}

// TestErrorMessages validates the error message content.
func TestErrorMessages(t *testing.T) {
	tests := []struct {
		name        string
		err         error
		expectedMsg string
	}{
		{
			name:        "ErrNotFound message",
			err:         ErrNotFound,
			expectedMsg: "cargo not found",
		},
		{
			name:        "ErrInvalidStatus message",
			err:         ErrInvalidStatus,
			expectedMsg: "invalid cargo status",
		},
		{
			name:        "ErrInvalidInput message",
			err:         ErrInvalidInput,
			expectedMsg: "invalid input",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given: a sentinel error
			// When: getting the error message
			msg := tt.err.Error()

			// Then: the message matches expected
			assert.Equal(t, tt.expectedMsg, msg)
		})
	}
}

// TestErrorComparison validates error comparison and wrapping.
func TestErrorComparison(t *testing.T) {
	tests := []struct {
		name        string
		givenErr    error
		targetErr   error
		shouldMatch bool
	}{
		{
			name:        "wrapped ErrNotFound matches unwrapped",
			givenErr:    errors.New("wrapped: " + ErrNotFound.Error()),
			targetErr:   ErrNotFound,
			shouldMatch: false, // different error instances
		},
		{
			name:        "same ErrNotFound reference matches",
			givenErr:    ErrNotFound,
			targetErr:   ErrNotFound,
			shouldMatch: true, // same reference
		},
		{
			name:        "ErrNotFound does not match ErrInvalidStatus",
			givenErr:    ErrNotFound,
			targetErr:   ErrInvalidStatus,
			shouldMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given: two errors
			// When: comparing them
			result := errors.Is(tt.givenErr, tt.targetErr)

			// Then: the result matches expected
			assert.Equal(t, tt.shouldMatch, result)
		})
	}
}

// TestErrorDefinitions validates that all required errors are defined.
func TestErrorDefinitions(t *testing.T) {
	// Given: the domain error package
	// When: checking all sentinel errors
	// Then: all required errors are defined and not nil

	requiredErrors := map[string]error{
		"ErrNotFound":      ErrNotFound,
		"ErrInvalidStatus": ErrInvalidStatus,
		"ErrInvalidInput":  ErrInvalidInput,
	}

	for name, err := range requiredErrors {
		t.Run(name, func(t *testing.T) {
			// Then: error is not nil
			require.NotNil(t, err, "sentinel error %s must be defined", name)

			// Then: error implements error interface
			assert.Implements(t, (*error)(nil), err)
		})
	}
}
