package vessel

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSentinelErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
		msg  string
	}{
		{"ErrNotFound", ErrNotFound, "vessel not found"},
		{"ErrInvalidInput", ErrInvalidInput, "invalid input"},
		{"ErrCapacityExceeded", ErrCapacityExceeded, "vessel capacity exceeded"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NotNil(t, tt.err)
			assert.Equal(t, tt.msg, tt.err.Error())
			assert.True(t, errors.Is(tt.err, tt.err))
		})
	}
}

func TestErrorsAreDistinct(t *testing.T) {
	assert.False(t, errors.Is(ErrNotFound, ErrInvalidInput))
	assert.False(t, errors.Is(ErrNotFound, ErrCapacityExceeded))
	assert.False(t, errors.Is(ErrInvalidInput, ErrCapacityExceeded))
}
