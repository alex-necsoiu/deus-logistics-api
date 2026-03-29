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

// TestCargoUpdateStatus tests the domain-level status transition enforcement
func TestCargoUpdateStatus(t *testing.T) {
	tests := []struct {
		name          string
		initialStatus CargoStatus
		targetStatus  CargoStatus
		shouldSucceed bool
		expectedErr   error
	}{
		// Valid transitions
		{
			name:          "pending to in_transit (valid transition)",
			initialStatus: CargoStatusPending,
			targetStatus:  CargoStatusInTransit,
			shouldSucceed: true,
			expectedErr:   nil,
		},
		{
			name:          "in_transit to delivered (valid transition)",
			initialStatus: CargoStatusInTransit,
			targetStatus:  CargoStatusDelivered,
			shouldSucceed: true,
			expectedErr:   nil,
		},

		// Invalid same-state transitions
		{
			name:          "pending to pending (same state)",
			initialStatus: CargoStatusPending,
			targetStatus:  CargoStatusPending,
			shouldSucceed: false,
			expectedErr:   ErrInvalidTransition,
		},
		{
			name:          "in_transit to in_transit (same state)",
			initialStatus: CargoStatusInTransit,
			targetStatus:  CargoStatusInTransit,
			shouldSucceed: false,
			expectedErr:   ErrInvalidTransition,
		},
		{
			name:          "delivered to delivered (same state)",
			initialStatus: CargoStatusDelivered,
			targetStatus:  CargoStatusDelivered,
			shouldSucceed: false,
			expectedErr:   ErrInvalidTransition,
		},

		// Invalid forward-skipping transitions
		{
			name:          "pending to delivered (skip in_transit)",
			initialStatus: CargoStatusPending,
			targetStatus:  CargoStatusDelivered,
			shouldSucceed: false,
			expectedErr:   ErrInvalidTransition,
		},

		// Invalid backward transitions
		{
			name:          "in_transit to pending (backward transition)",
			initialStatus: CargoStatusInTransit,
			targetStatus:  CargoStatusPending,
			shouldSucceed: false,
			expectedErr:   ErrInvalidTransition,
		},
		{
			name:          "delivered to pending (backward transition)",
			initialStatus: CargoStatusDelivered,
			targetStatus:  CargoStatusPending,
			shouldSucceed: false,
			expectedErr:   ErrInvalidTransition,
		},
		{
			name:          "delivered to in_transit (backward transition)",
			initialStatus: CargoStatusDelivered,
			targetStatus:  CargoStatusInTransit,
			shouldSucceed: false,
			expectedErr:   ErrInvalidTransition,
		},

		// Terminal state testing
		{
			name:          "delivered blocks all transitions",
			initialStatus: CargoStatusDelivered,
			targetStatus:  CargoStatusDelivered,
			shouldSucceed: false,
			expectedErr:   ErrInvalidTransition,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a cargo with the initial status
			cargo := &Cargo{
				ID:        uuid.New(),
				Status:    tt.initialStatus,
				Name:      "Test Cargo",
				Weight:    100.0,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			originalUpdatedAt := cargo.UpdatedAt
			time.Sleep(1 * time.Millisecond) // Ensure timestamp difference if transition succeeds

			// Attempt the transition
			err := cargo.UpdateStatus(tt.targetStatus)

			if tt.shouldSucceed {
				require.NoError(t, err, "expected transition to succeed")
				assert.Equal(t, tt.targetStatus, cargo.Status, "status should be updated")
				assert.True(t, cargo.UpdatedAt.After(originalUpdatedAt), "UpdatedAt should be newer after valid transition")
			} else {
				require.Error(t, err, "expected transition to fail")
				assert.ErrorIs(t, err, tt.expectedErr, "expected error type should match")
				assert.Equal(t, tt.initialStatus, cargo.Status, "status should not change after failed transition")
				assert.Equal(t, originalUpdatedAt, cargo.UpdatedAt, "UpdatedAt should not change after failed transition")
			}
		})
	}
}

// TestCargoCanTransitionTo tests the read-only transition validation
func TestCargoCanTransitionTo(t *testing.T) {
	tests := []struct {
		name          string
		initialStatus CargoStatus
		targetStatus  CargoStatus
		canTransition bool
	}{
		// Valid transitions
		{
			name:          "pending can transition to in_transit",
			initialStatus: CargoStatusPending,
			targetStatus:  CargoStatusInTransit,
			canTransition: true,
		},
		{
			name:          "in_transit can transition to delivered",
			initialStatus: CargoStatusInTransit,
			targetStatus:  CargoStatusDelivered,
			canTransition: true,
		},

		// Invalid transitions
		{
			name:          "pending cannot transition to pending",
			initialStatus: CargoStatusPending,
			targetStatus:  CargoStatusPending,
			canTransition: false,
		},
		{
			name:          "pending cannot transition to delivered",
			initialStatus: CargoStatusPending,
			targetStatus:  CargoStatusDelivered,
			canTransition: false,
		},
		{
			name:          "in_transit cannot transition to pending",
			initialStatus: CargoStatusInTransit,
			targetStatus:  CargoStatusPending,
			canTransition: false,
		},
		{
			name:          "in_transit cannot transition to in_transit",
			initialStatus: CargoStatusInTransit,
			targetStatus:  CargoStatusInTransit,
			canTransition: false,
		},
		{
			name:          "delivered cannot transition to pending",
			initialStatus: CargoStatusDelivered,
			targetStatus:  CargoStatusPending,
			canTransition: false,
		},
		{
			name:          "delivered cannot transition to in_transit",
			initialStatus: CargoStatusDelivered,
			targetStatus:  CargoStatusInTransit,
			canTransition: false,
		},
		{
			name:          "delivered cannot transition to delivered",
			initialStatus: CargoStatusDelivered,
			targetStatus:  CargoStatusDelivered,
			canTransition: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a cargo with the initial status
			cargo := &Cargo{
				ID:        uuid.New(),
				Status:    tt.initialStatus,
				Name:      "Test Cargo",
				Weight:    100.0,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}

			originalStatus := cargo.Status
			originalUpdatedAt := cargo.UpdatedAt

			// Check if transition is possible (read-only operation)
			result := cargo.CanTransitionTo(tt.targetStatus)

			// Verify result
			assert.Equal(t, tt.canTransition, result, "CanTransitionTo result should match expected")

			// Verify state is not modified (read-only check)
			assert.Equal(t, originalStatus, cargo.Status, "CanTransitionTo should not modify status")
			assert.Equal(t, originalUpdatedAt, cargo.UpdatedAt, "CanTransitionTo should not modify UpdatedAt")
		})
	}
}

// TestCargoTransitionSequence tests a complete lifecycle transition chain
func TestCargoTransitionSequence(t *testing.T) {
	// Create a cargo starting in pending status
	cargo := &Cargo{
		ID:        uuid.New(),
		Status:    CargoStatusPending,
		Name:      "Test Cargo",
		Weight:    100.0,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Verify initial state
	assert.Equal(t, CargoStatusPending, cargo.Status)

	// Transition: pending → in_transit
	err := cargo.UpdateStatus(CargoStatusInTransit)
	require.NoError(t, err)
	assert.Equal(t, CargoStatusInTransit, cargo.Status)

	// Verify pending→delivered is now blocked
	err = cargo.UpdateStatus(CargoStatusDelivered)
	require.NoError(t, err) // Now that we're in_transit, this should succeed
	assert.Equal(t, CargoStatusDelivered, cargo.Status)

	// Verify delivered is terminal
	err = cargo.UpdateStatus(CargoStatusInTransit)
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidTransition)
	assert.Equal(t, CargoStatusDelivered, cargo.Status, "status should remain delivered after failed transition")
}
