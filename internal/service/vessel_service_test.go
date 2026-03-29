package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alex-necsoiu/deus-logistics-api/internal/domain/vessel"
)

// mockVesselRepository is a mock implementation of vessel.Repository for testing.
type mockVesselRepository struct {
	createFunc       func(ctx context.Context, input vessel.CreateVesselInput) (*vessel.Vessel, error)
	getByIDFunc      func(ctx context.Context, id uuid.UUID) (*vessel.Vessel, error)
	listFunc         func(ctx context.Context) ([]*vessel.Vessel, error)
	updateLocFunc    func(ctx context.Context, id uuid.UUID, location string) (*vessel.Vessel, error)
	updateCapFunc    func(ctx context.Context, id uuid.UUID, capacity float64) (*vessel.Vessel, error)
}

func (m *mockVesselRepository) Create(ctx context.Context, input vessel.CreateVesselInput) (*vessel.Vessel, error) {
	return m.createFunc(ctx, input)
}

func (m *mockVesselRepository) GetByID(ctx context.Context, id uuid.UUID) (*vessel.Vessel, error) {
	return m.getByIDFunc(ctx, id)
}

func (m *mockVesselRepository) List(ctx context.Context) ([]*vessel.Vessel, error) {
	return m.listFunc(ctx)
}

func (m *mockVesselRepository) UpdateLocation(ctx context.Context, id uuid.UUID, location string) (*vessel.Vessel, error) {
	return m.updateLocFunc(ctx, id, location)
}

func (m *mockVesselRepository) UpdateCapacity(ctx context.Context, id uuid.UUID, capacity float64) (*vessel.Vessel, error) {
	return m.updateCapFunc(ctx, id, capacity)
}

// TestVesselServiceCreateVessel tests vessel creation with valid input.
func TestVesselServiceCreateVessel(t *testing.T) {
	// Given
	ctx := context.Background()

	mockRepo := &mockVesselRepository{
		createFunc: func(ctx context.Context, input vessel.CreateVesselInput) (*vessel.Vessel, error) {
			assert.Equal(t, "Ship Alpha", input.Name)
			assert.Equal(t, 5000.0, input.Capacity)
			assert.Equal(t, "Hamburg", input.CurrentLocation)

			return &vessel.Vessel{
				ID:              uuid.New(),
				Name:            input.Name,
				Capacity:        input.Capacity,
				CurrentLocation: input.CurrentLocation,
			}, nil
		},
	}

	service := NewVesselService(mockRepo)

	// When
	input := vessel.CreateVesselInput{
		Name:            "Ship Alpha",
		Capacity:        5000.0,
		CurrentLocation: "Hamburg",
	}
	result, err := service.CreateVessel(ctx, input)

	// Then
	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Ship Alpha", result.Name)
	assert.Equal(t, 5000.0, result.Capacity)
}


// TestVesselServiceGetVessel tests retrieving vessel by ID.
func TestVesselServiceGetVessel(t *testing.T) {
	// Given
	ctx := context.Background()
	vesselID := uuid.New()

	mockRepo := &mockVesselRepository{
		getByIDFunc: func(ctx context.Context, id uuid.UUID) (*vessel.Vessel, error) {
			assert.Equal(t, vesselID, id)
			return &vessel.Vessel{
				ID:   vesselID,
				Name: "Test Vessel",
			}, nil
		},
	}

	service := NewVesselService(mockRepo)

	// When
	result, err := service.GetVessel(ctx, vesselID)

	// Then
	require.NoError(t, err)
	assert.Equal(t, vesselID, result.ID)
	assert.Equal(t, "Test Vessel", result.Name)
}

// TestVesselServiceGetVesselNotFound tests retrieving non-existent vessel.
func TestVesselServiceGetVesselNotFound(t *testing.T) {
	// Given
	ctx := context.Background()
	vesselID := uuid.New()

	mockRepo := &mockVesselRepository{
		getByIDFunc: func(ctx context.Context, id uuid.UUID) (*vessel.Vessel, error) {
			return nil, vessel.ErrNotFound
		},
	}

	service := NewVesselService(mockRepo)

	// When
	result, err := service.GetVessel(ctx, vesselID)

	// Then
	assert.Nil(t, result)
	assert.True(t, errors.Is(err, vessel.ErrNotFound))
}

// TestVesselServiceListVessels tests listing all vessels.
func TestVesselServiceListVessels(t *testing.T) {
	// Given
	ctx := context.Background()

	mockRepo := &mockVesselRepository{
		listFunc: func(ctx context.Context) ([]*vessel.Vessel, error) {
			return []*vessel.Vessel{
				{ID: uuid.New(), Name: "Vessel1"},
				{ID: uuid.New(), Name: "Vessel2"},
			}, nil
		},
	}

	service := NewVesselService(mockRepo)

	// When
	result, err := service.ListVessels(ctx)

	// Then
	require.NoError(t, err)
	assert.Len(t, result, 2)
}

// TestVesselServiceUpdateLocation tests updating vessel location.
func TestVesselServiceUpdateLocation(t *testing.T) {
	// Given
	ctx := context.Background()
	vesselID := uuid.New()

	mockRepo := &mockVesselRepository{
		updateLocFunc: func(ctx context.Context, id uuid.UUID, location string) (*vessel.Vessel, error) {
			assert.Equal(t, "Rotterdam", location)
			return &vessel.Vessel{
				ID:              id,
				Name:            "Test Vessel",
				CurrentLocation: location,
			}, nil
		},
	}

	service := NewVesselService(mockRepo)

	// When
	result, err := service.UpdateVesselLocation(ctx, vesselID, "Rotterdam")

	// Then
	require.NoError(t, err)
	assert.Equal(t, "Rotterdam", result.CurrentLocation)
}
