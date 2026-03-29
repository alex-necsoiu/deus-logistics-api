package cargo

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCargoStatusIsValid(t *testing.T) {
	tests := []struct {
		name     string
		status   CargoStatus
		expected bool
	}{
		{"valid: pending", CargoStatusPending, true},
		{"valid: in_transit", CargoStatusInTransit, true},
		{"valid: delivered", CargoStatusDelivered, true},
		{"invalid: unknown", CargoStatus("unknown"), false},
		{"invalid: empty", CargoStatus(""), false},
		{"invalid: typo", CargoStatus("in-transit"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.status.IsValid())
		})
	}
}

func TestCargoStatusString(t *testing.T) {
	tests := []struct {
		name     string
		status   CargoStatus
		expected string
	}{
		{"pending", CargoStatusPending, "pending"},
		{"in_transit", CargoStatusInTransit, "in_transit"},
		{"delivered", CargoStatusDelivered, "delivered"},
		{"custom", CargoStatus("custom"), "custom"},
		{"empty", CargoStatus(""), ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.status.String())
		})
	}
}

func TestCargoIsDelivered(t *testing.T) {
	tests := []struct {
		name     string
		status   CargoStatus
		expected bool
	}{
		{"delivered returns true", CargoStatusDelivered, true},
		{"pending returns false", CargoStatusPending, false},
		{"in_transit returns false", CargoStatusInTransit, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Cargo{
				ID:        uuid.New(),
				Name:      "test",
				Status:    tt.status,
				VesselID:  uuid.New(),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
			assert.Equal(t, tt.expected, c.IsDelivered())
		})
	}
}

func TestCargoIsInTransit(t *testing.T) {
	tests := []struct {
		name     string
		status   CargoStatus
		expected bool
	}{
		{"in_transit returns true", CargoStatusInTransit, true},
		{"pending returns false", CargoStatusPending, false},
		{"delivered returns false", CargoStatusDelivered, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Cargo{
				ID:        uuid.New(),
				Name:      "test",
				Status:    tt.status,
				VesselID:  uuid.New(),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
			assert.Equal(t, tt.expected, c.IsInTransit())
		})
	}
}

func TestCreateCargoInputValidation(t *testing.T) {
	vesselID := uuid.New()

	tests := []struct {
		name  string
		input CreateCargoInput
		valid bool
	}{
		{
			name:  "valid input",
			input: CreateCargoInput{Name: "Electronics", Description: "Consumer electronics", Weight: 1500.0, VesselID: vesselID},
			valid: true,
		},
		{
			name:  "valid without description",
			input: CreateCargoInput{Name: "Raw Materials", Weight: 5000.0, VesselID: vesselID},
			valid: true,
		},
		{
			name:  "empty name",
			input: CreateCargoInput{Name: "", Weight: 100.0, VesselID: vesselID},
			valid: false,
		},
		{
			name:  "zero weight",
			input: CreateCargoInput{Name: "Test", Weight: 0.0, VesselID: vesselID},
			valid: false,
		},
		{
			name:  "negative weight",
			input: CreateCargoInput{Name: "Test", Weight: -100.0, VesselID: vesselID},
			valid: false,
		},
		{
			name:  "zero UUID vessel",
			input: CreateCargoInput{Name: "Test", Weight: 100.0, VesselID: uuid.UUID{}},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.input.Validate()
			if tt.valid {
				require.NoError(t, err)
				return
			}
			require.Error(t, err)
			assert.ErrorIs(t, err, ErrInvalidInput)
		})
	}
}
