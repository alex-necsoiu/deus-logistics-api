package vessel

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVesselStruct(t *testing.T) {
	id := uuid.New()
	now := time.Now()

	v := &Vessel{
		ID:              id,
		Name:            "MS Pacific",
		Capacity:        50000.0,
		CurrentLocation: "Singapore",
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	assert.Equal(t, id, v.ID)
	assert.Equal(t, "MS Pacific", v.Name)
	assert.Equal(t, 50000.0, v.Capacity)
	assert.Equal(t, "Singapore", v.CurrentLocation)
	assert.Equal(t, now, v.CreatedAt)
	assert.Equal(t, now, v.UpdatedAt)
}

func TestCanCargoFit(t *testing.T) {
	v := &Vessel{
		ID:       uuid.New(),
		Name:     "MS Pacific",
		Capacity: 50000.0,
	}

	tests := []struct {
		name     string
		weight   float64
		expected bool
	}{
		{"weight within capacity", 30000.0, true},
		{"weight at exact capacity", 50000.0, true},
		{"weight exceeds capacity", 50001.0, false},
		{"zero weight", 0, false},
		{"negative weight", -100.0, false},
		{"small weight", 0.01, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := v.CanCargoFit(tt.weight)
			assert.Equal(t, tt.expected, got)
		})
	}
}

func TestCreateVesselInputValidate(t *testing.T) {
	tests := []struct {
		name    string
		input   CreateVesselInput
		wantErr bool
	}{
		{
			name:    "valid input",
			input:   CreateVesselInput{Name: "MS Pacific", Capacity: 50000, CurrentLocation: "Singapore"},
			wantErr: false,
		},
		{
			name:    "empty name",
			input:   CreateVesselInput{Name: "", Capacity: 50000, CurrentLocation: "Singapore"},
			wantErr: true,
		},
		{
			name:    "zero capacity",
			input:   CreateVesselInput{Name: "MS Pacific", Capacity: 0, CurrentLocation: "Singapore"},
			wantErr: true,
		},
		{
			name:    "negative capacity",
			input:   CreateVesselInput{Name: "MS Pacific", Capacity: -1000, CurrentLocation: "Singapore"},
			wantErr: true,
		},
		{
			name:    "empty location",
			input:   CreateVesselInput{Name: "MS Pacific", Capacity: 50000, CurrentLocation: ""},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.input.Validate()
			if tt.wantErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, ErrInvalidInput)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
