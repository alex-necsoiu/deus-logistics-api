package tracking

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
		{"ErrInvalidEntry", ErrInvalidEntry, "invalid tracking entry"},
		{"ErrCargoNotFound", ErrCargoNotFound, "cargo not found"},
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
	assert.False(t, errors.Is(ErrInvalidEntry, ErrCargoNotFound))
}
