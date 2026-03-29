package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alex-necsoiu/deus-logistics-api/internal/domain/tracking"
)

// mockTrackingRepositoryImpl is a full mock implementation of tracking.Repository for testing.
type mockTrackingRepositoryImpl struct {
	createFunc      func(ctx context.Context, input tracking.AddTrackingInput) (*tracking.TrackingEntry, error)
	listByCargoFunc func(ctx context.Context, cargoID uuid.UUID) ([]*tracking.TrackingEntry, error)
}

func (m *mockTrackingRepositoryImpl) Create(ctx context.Context, input tracking.AddTrackingInput) (*tracking.TrackingEntry, error) {
	return m.createFunc(ctx, input)
}

func (m *mockTrackingRepositoryImpl) ListByCargoID(ctx context.Context, cargoID uuid.UUID) ([]*tracking.TrackingEntry, error) {
	return m.listByCargoFunc(ctx, cargoID)
}

// TestTrackingServiceAddEntry tests adding a tracking entry.
func TestTrackingServiceAddEntry(t *testing.T) {
	// Given
	ctx := context.Background()
	cargoID := uuid.New()

	mockRepo := &mockTrackingRepositoryImpl{
		createFunc: func(ctx context.Context, input tracking.AddTrackingInput) (*tracking.TrackingEntry, error) {
			assert.Equal(t, cargoID, input.CargoID)
			assert.Equal(t, "Port of Hamburg", input.Location)
			return &tracking.TrackingEntry{
				ID:       uuid.New(),
				CargoID:  cargoID,
				Location: input.Location,
			}, nil
		},
	}

	service := NewTrackingService(mockRepo)

	// When
	input := tracking.AddTrackingInput{
		CargoID:  cargoID,
		Location: "Port of Hamburg",
		Status:   "in_transit",
		Note:     "Cargo loaded onto vessel",
	}
	result, err := service.AddTrackingEntry(ctx, input)

	// Then
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, cargoID, result.CargoID)
	assert.Equal(t, "Port of Hamburg", result.Location)
}

// TestTrackingServiceGetHistory tests retrieving tracking history.
func TestTrackingServiceGetHistory(t *testing.T) {
	// Given
	ctx := context.Background()
	cargoID := uuid.New()

	mockRepo := &mockTrackingRepositoryImpl{
		listByCargoFunc: func(ctx context.Context, id uuid.UUID) ([]*tracking.TrackingEntry, error) {
			assert.Equal(t, cargoID, id)
			return []*tracking.TrackingEntry{
				{ID: uuid.New(), CargoID: cargoID, Location: "Hamburg"},
				{ID: uuid.New(), CargoID: cargoID, Location: "Rotterdam"},
			}, nil
		},
	}

	service := NewTrackingService(mockRepo)

	// When
	result, err := service.GetTrackingHistory(ctx, cargoID)

	// Then
	require.NoError(t, err)
	assert.Len(t, result, 2)
}
