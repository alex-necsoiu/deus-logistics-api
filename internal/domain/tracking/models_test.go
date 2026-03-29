package tracking

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTrackingEntryStruct(t *testing.T) {
	id := uuid.New()
	cargoID := uuid.New()
	ts := time.Now()

	entry := &TrackingEntry{
		ID:        id,
		CargoID:   cargoID,
		Location:  "Rotterdam",
		Status:    "in_transit",
		Note:      "Arrived at port",
		Timestamp: ts,
	}

	assert.Equal(t, id, entry.ID)
	assert.Equal(t, cargoID, entry.CargoID)
	assert.Equal(t, "Rotterdam", entry.Location)
	assert.Equal(t, "in_transit", entry.Status)
	assert.Equal(t, "Arrived at port", entry.Note)
	assert.Equal(t, ts, entry.Timestamp)
}

func TestAddTrackingInputValidation(t *testing.T) {
	cargoID := uuid.New()

	tests := []struct {
		name  string
		input AddTrackingInput
		valid bool
	}{
		{
			name:  "valid input",
			input: AddTrackingInput{CargoID: cargoID, Location: "Rotterdam", Status: "in_transit", Note: "Arrived"},
			valid: true,
		},
		{
			name:  "valid without note",
			input: AddTrackingInput{CargoID: cargoID, Location: "Rotterdam", Status: "in_transit"},
			valid: true,
		},
		{
			name:  "zero UUID cargo",
			input: AddTrackingInput{CargoID: uuid.UUID{}, Location: "Rotterdam", Status: "in_transit"},
			valid: false,
		},
		{
			name:  "empty location",
			input: AddTrackingInput{CargoID: cargoID, Location: "", Status: "in_transit"},
			valid: false,
		},
		{
			name:  "empty status",
			input: AddTrackingInput{CargoID: cargoID, Location: "Rotterdam", Status: ""},
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
			assert.ErrorIs(t, err, ErrInvalidEntry)
		})
	}
}
