package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alex-necsoiu/deus-logistics-api/internal/domain/cargo"
	"github.com/alex-necsoiu/deus-logistics-api/internal/domain/tracking"
)

// --- Mocks ---

type mockCargoRepository struct {
	createFunc       func(ctx context.Context, input cargo.CreateCargoInput) (*cargo.Cargo, error)
	getByIDFunc      func(ctx context.Context, id uuid.UUID) (*cargo.Cargo, error)
	listFunc         func(ctx context.Context) ([]*cargo.Cargo, error)
	listByVesselFunc func(ctx context.Context, vesselID uuid.UUID) ([]*cargo.Cargo, error)
	updateStatusFunc func(ctx context.Context, id uuid.UUID, status cargo.CargoStatus) (*cargo.Cargo, error)
}

func (m *mockCargoRepository) Create(ctx context.Context, input cargo.CreateCargoInput) (*cargo.Cargo, error) {
	return m.createFunc(ctx, input)
}

func (m *mockCargoRepository) GetByID(ctx context.Context, id uuid.UUID) (*cargo.Cargo, error) {
	return m.getByIDFunc(ctx, id)
}

func (m *mockCargoRepository) List(ctx context.Context) ([]*cargo.Cargo, error) {
	return m.listFunc(ctx)
}

func (m *mockCargoRepository) ListByVesselID(ctx context.Context, vesselID uuid.UUID) ([]*cargo.Cargo, error) {
	return m.listByVesselFunc(ctx, vesselID)
}

func (m *mockCargoRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status cargo.CargoStatus) (*cargo.Cargo, error) {
	return m.updateStatusFunc(ctx, id, status)
}

type mockTrackingRepository struct {
	appendFunc func(ctx context.Context, input tracking.AddTrackingInput) (*tracking.TrackingEntry, error)
}

func (m *mockTrackingRepository) Append(ctx context.Context, input tracking.AddTrackingInput) (*tracking.TrackingEntry, error) {
	return m.appendFunc(ctx, input)
}

func (m *mockTrackingRepository) ListByCargoID(_ context.Context, _ uuid.UUID) ([]*tracking.TrackingEntry, error) {
	return nil, nil
}

type mockEventPublisher struct {
	publishFunc func(ctx context.Context, event cargo.StatusChangedEvent) error
}

func (m *mockEventPublisher) PublishStatusChanged(ctx context.Context, event cargo.StatusChangedEvent) error {
	return m.publishFunc(ctx, event)
}

// --- Tests ---

func TestCargoServiceCreateCargo(t *testing.T) {
	// Given
	ctx := context.Background()
	vesselID := uuid.New()

	mockRepo := &mockCargoRepository{
		createFunc: func(_ context.Context, input cargo.CreateCargoInput) (*cargo.Cargo, error) {
			assert.Equal(t, "Electronics", input.Name)
			assert.Equal(t, "Fragile goods", input.Description)
			assert.Equal(t, 100.0, input.Weight)
			assert.Equal(t, vesselID, input.VesselID)
			return &cargo.Cargo{
				ID:          uuid.New(),
				Name:        input.Name,
				Description: input.Description,
				Weight:      input.Weight,
				Status:      cargo.CargoStatusPending,
				VesselID:    vesselID,
			}, nil
		},
	}
	svc := NewCargoService(mockRepo, nil, nil)

	// When
	input := cargo.CreateCargoInput{
		Name:        "Electronics",
		Description: "Fragile goods",
		Weight:      100.0,
		VesselID:    vesselID,
	}
	result, err := svc.CreateCargo(ctx, input)

	// Then
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Electronics", result.Name)
	assert.Equal(t, cargo.CargoStatusPending, result.Status)
}

func TestCargoServiceGetCargo(t *testing.T) {
	// Given
	ctx := context.Background()
	cargoID := uuid.New()

	mockRepo := &mockCargoRepository{
		getByIDFunc: func(_ context.Context, id uuid.UUID) (*cargo.Cargo, error) {
			assert.Equal(t, cargoID, id)
			return &cargo.Cargo{
				ID:     cargoID,
				Name:   "Test Cargo",
				Status: cargo.CargoStatusPending,
			}, nil
		},
	}
	svc := NewCargoService(mockRepo, nil, nil)

	// When
	result, err := svc.GetCargo(ctx, cargoID)

	// Then
	require.NoError(t, err)
	assert.Equal(t, cargoID, result.ID)
	assert.Equal(t, "Test Cargo", result.Name)
}

func TestCargoServiceGetCargoNotFound(t *testing.T) {
	// Given
	ctx := context.Background()
	cargoID := uuid.New()

	mockRepo := &mockCargoRepository{
		getByIDFunc: func(_ context.Context, _ uuid.UUID) (*cargo.Cargo, error) {
			return nil, cargo.ErrNotFound
		},
	}
	svc := NewCargoService(mockRepo, nil, nil)

	// When
	result, err := svc.GetCargo(ctx, cargoID)

	// Then
	require.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, errors.Is(err, cargo.ErrNotFound))
}

func TestCargoServiceListCargoes(t *testing.T) {
	// Given
	ctx := context.Background()

	mockRepo := &mockCargoRepository{
		listFunc: func(_ context.Context) ([]*cargo.Cargo, error) {
			return []*cargo.Cargo{
				{ID: uuid.New(), Name: "Cargo1"},
				{ID: uuid.New(), Name: "Cargo2"},
			}, nil
		},
	}
	svc := NewCargoService(mockRepo, nil, nil)

	// When
	result, err := svc.ListCargoes(ctx)

	// Then
	require.NoError(t, err)
	assert.Len(t, result, 2)
}

func TestCargoServiceListCargoByVesselID(t *testing.T) {
	// Given
	ctx := context.Background()
	vesselID := uuid.New()

	mockRepo := &mockCargoRepository{
		listByVesselFunc: func(_ context.Context, vid uuid.UUID) ([]*cargo.Cargo, error) {
			assert.Equal(t, vesselID, vid)
			return []*cargo.Cargo{
				{ID: uuid.New(), Name: "Cargo1", VesselID: vesselID},
			}, nil
		},
	}
	svc := NewCargoService(mockRepo, nil, nil)

	// When
	result, err := svc.ListCargoByVesselID(ctx, vesselID)

	// Then
	require.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, vesselID, result[0].VesselID)
}

func TestCargoServiceUpdateCargoStatus(t *testing.T) {
	// Given
	ctx := context.Background()
	cargoID := uuid.New()

	mockRepo := &mockCargoRepository{
		getByIDFunc: func(_ context.Context, id uuid.UUID) (*cargo.Cargo, error) {
			return &cargo.Cargo{
				ID:     id,
				Name:   "Test Cargo",
				Status: cargo.CargoStatusPending,
			}, nil
		},
		updateStatusFunc: func(_ context.Context, id uuid.UUID, status cargo.CargoStatus) (*cargo.Cargo, error) {
			return &cargo.Cargo{
				ID:     id,
				Name:   "Test Cargo",
				Status: status,
			}, nil
		},
	}

	mockPublisher := &mockEventPublisher{
		publishFunc: func(_ context.Context, event cargo.StatusChangedEvent) error {
			assert.Equal(t, cargoID.String(), event.CargoID)
			assert.Equal(t, cargo.CargoStatusPending, event.OldStatus)
			assert.Equal(t, cargo.CargoStatusInTransit, event.NewStatus)
			return nil
		},
	}

	mockTracker := &mockTrackingRepository{
			appendFunc: func(_ context.Context, _ tracking.AddTrackingInput) (*tracking.TrackingEntry, error) {
			return &tracking.TrackingEntry{ID: uuid.New()}, nil
		},
	}

	svc := NewCargoService(mockRepo, mockPublisher, mockTracker)

	// When
	result, err := svc.UpdateCargoStatus(ctx, cargoID, cargo.CargoStatusInTransit)

	// Then
	require.NoError(t, err)
	assert.Equal(t, cargo.CargoStatusInTransit, result.Status)
}

func TestCargoServiceUpdateCargoStatusInvalidStatus(t *testing.T) {
	// Given
	ctx := context.Background()
	cargoID := uuid.New()
	svc := NewCargoService(nil, nil, nil)

	// When
	result, err := svc.UpdateCargoStatus(ctx, cargoID, cargo.CargoStatus("invalid"))

	// Then
	require.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, errors.Is(err, cargo.ErrInvalidStatus))
}

func TestCargoServiceUpdateCargoStatusRepoError(t *testing.T) {
	// Given
	ctx := context.Background()
	cargoID := uuid.New()

	mockRepo := &mockCargoRepository{
		getByIDFunc: func(_ context.Context, _ uuid.UUID) (*cargo.Cargo, error) {
			return nil, cargo.ErrNotFound
		},
	}
	svc := NewCargoService(mockRepo, nil, nil)

	// When
	result, err := svc.UpdateCargoStatus(ctx, cargoID, cargo.CargoStatusDelivered)

	// Then
	require.Error(t, err)
	assert.Nil(t, result)
	assert.True(t, errors.Is(err, cargo.ErrNotFound))
}
