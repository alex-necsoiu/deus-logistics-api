package cargo

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestCargoEventStruct(t *testing.T) {
	id := uuid.New()
	cargoID := uuid.New()
	ts := time.Now()

	event := &CargoEvent{
		ID:        id,
		CargoID:   cargoID,
		OldStatus: CargoStatusPending,
		NewStatus: CargoStatusInTransit,
		Timestamp: ts,
	}

	assert.Equal(t, id, event.ID)
	assert.Equal(t, cargoID, event.CargoID)
	assert.Equal(t, CargoStatusPending, event.OldStatus)
	assert.Equal(t, CargoStatusInTransit, event.NewStatus)
	assert.Equal(t, ts, event.Timestamp)
}

func TestStatusChangedEventStruct(t *testing.T) {
	id := uuid.New().String()
	cargoID := uuid.New().String()
	ts := time.Now()

	event := &StatusChangedEvent{
		ID:        id,
		EventType: "cargo.status_changed",
		CargoID:   cargoID,
		OldStatus: CargoStatusPending,
		NewStatus: CargoStatusInTransit,
		Timestamp: ts,
	}

	assert.Equal(t, id, event.ID)
	assert.Equal(t, "cargo.status_changed", event.EventType)
	assert.Equal(t, cargoID, event.CargoID)
	assert.Equal(t, CargoStatusPending, event.OldStatus)
	assert.Equal(t, CargoStatusInTransit, event.NewStatus)
	assert.Equal(t, ts, event.Timestamp)
}

func TestStatusChangedEventStatusValues(t *testing.T) {
	tests := []struct {
		name      string
		oldStatus CargoStatus
		newStatus CargoStatus
	}{
		{"pending to in_transit", CargoStatusPending, CargoStatusInTransit},
		{"in_transit to delivered", CargoStatusInTransit, CargoStatusDelivered},
		{"pending to delivered", CargoStatusPending, CargoStatusDelivered},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := StatusChangedEvent{
				ID:        uuid.New().String(),
				CargoID:   uuid.New().String(),
				OldStatus: tt.oldStatus,
				NewStatus: tt.newStatus,
				Timestamp: time.Now(),
			}
			assert.True(t, event.OldStatus.IsValid())
			assert.True(t, event.NewStatus.IsValid())
			assert.NotEqual(t, event.OldStatus, event.NewStatus)
		})
	}
}
