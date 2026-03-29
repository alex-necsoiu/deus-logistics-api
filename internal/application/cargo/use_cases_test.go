package cargo

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	domaincargo "github.com/alex-necsoiu/deus-logistics-api/internal/domain/cargo"
)

// MockCargoRepository is a mock implementation of CargoRepository
type MockCargoRepository struct {
	mock.Mock
}

func (m *MockCargoRepository) Create(ctx context.Context, input domaincargo.CreateCargoInput) (*domaincargo.Cargo, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domaincargo.Cargo), args.Error(1)
}

func (m *MockCargoRepository) GetByID(ctx context.Context, id uuid.UUID) (*domaincargo.Cargo, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domaincargo.Cargo), args.Error(1)
}

func (m *MockCargoRepository) List(ctx context.Context) ([]*domaincargo.Cargo, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domaincargo.Cargo), args.Error(1)
}

func (m *MockCargoRepository) ListByVesselID(ctx context.Context, vesselID uuid.UUID) ([]*domaincargo.Cargo, error) {
	args := m.Called(ctx, vesselID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domaincargo.Cargo), args.Error(1)
}

func (m *MockCargoRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status domaincargo.CargoStatus) (*domaincargo.Cargo, error) {
	args := m.Called(ctx, id, status)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domaincargo.Cargo), args.Error(1)
}

// TestCreateCargoUseCase tests the CreateCargoUseCase
func TestCreateCargoUseCase(t *testing.T) {
	ctx := context.Background()
	vesselID := uuid.New()

	tests := []struct {
		name      string
		input     domaincargo.CreateCargoInput
		setupMock func(*MockCargoRepository)
		wantErr   bool
		errMsg    string
	}{
		{
			name: "successful cargo creation",
			input: domaincargo.CreateCargoInput{
				Name:        "Container ABC123",
				Description: "Electronics shipment",
				Weight:      500.0,
				VesselID:    vesselID,
			},
			setupMock: func(m *MockCargoRepository) {
				m.On("Create", ctx, mock.MatchedBy(func(input domaincargo.CreateCargoInput) bool {
					return input.Name == "Container ABC123"
				})).Return(&domaincargo.Cargo{
					ID:          uuid.New(),
					Name:        "Container ABC123",
					Description: "Electronics shipment",
					Weight:      500.0,
					Status:      domaincargo.CargoStatusPending,
					VesselID:    vesselID,
				}, nil)
			},
			wantErr: false,
		},
		{
			name: "validation error - empty name",
			input: domaincargo.CreateCargoInput{
				Name:        "",
				Description: "Test cargo",
				Weight:      100.0,
				VesselID:    vesselID,
			},
			setupMock: func(m *MockCargoRepository) {},
			wantErr:   true,
		},
		{
			name: "validation error - invalid weight",
			input: domaincargo.CreateCargoInput{
				Name:        "Test Cargo",
				Description: "Test cargo",
				Weight:      0, // Invalid: must be positive
				VesselID:    vesselID,
			},
			setupMock: func(m *MockCargoRepository) {},
			wantErr:   true,
		},
		{
			name: "repository error",
			input: domaincargo.CreateCargoInput{
				Name:        "Container XYZ",
				Description: "Test shipment",
				Weight:      250.0,
				VesselID:    vesselID,
			},
			setupMock: func(m *MockCargoRepository) {
				m.On("Create", ctx, mock.Anything).Return(nil, errors.New("database error"))
			},
			wantErr: true,
			errMsg:  "create_cargo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockCargoRepository)
			tt.setupMock(mockRepo)

			uc := NewCreateCargoUseCase(mockRepo)
			cargo, err := uc.Execute(ctx, tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, cargo)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, cargo)
				assert.Equal(t, tt.input.Name, cargo.Name)
				assert.Equal(t, tt.input.Weight, cargo.Weight)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

// TestGetCargoUseCase tests the GetCargoUseCase
func TestGetCargoUseCase(t *testing.T) {
	ctx := context.Background()
	cargoID := uuid.New()
	vesselID := uuid.New()

	tests := []struct {
		name      string
		cargoID   uuid.UUID
		setupMock func(*MockCargoRepository)
		wantErr   bool
	}{
		{
			name:    "successful get",
			cargoID: cargoID,
			setupMock: func(m *MockCargoRepository) {
				m.On("GetByID", ctx, cargoID).Return(&domaincargo.Cargo{
					ID:       cargoID,
					Name:     "Container ABC",
					Weight:   100.0,
					Status:   domaincargo.CargoStatusPending,
					VesselID: vesselID,
				}, nil)
			},
			wantErr: false,
		},
		{
			name:    "cargo not found",
			cargoID: cargoID,
			setupMock: func(m *MockCargoRepository) {
				m.On("GetByID", ctx, cargoID).Return(nil, errors.New("not found"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockCargoRepository)
			tt.setupMock(mockRepo)

			uc := NewGetCargoUseCase(mockRepo)
			cargo, err := uc.Execute(ctx, tt.cargoID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, cargo)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, cargo)
				assert.Equal(t, tt.cargoID, cargo.ID)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

// TestListCargosUseCase tests the ListCargosUseCase
func TestListCargosUseCase(t *testing.T) {
	ctx := context.Background()
	vesselID := uuid.New()

	tests := []struct {
		name      string
		setupMock func(*MockCargoRepository)
		wantErr   bool
		wantCount int
	}{
		{
			name: "successful list empty",
			setupMock: func(m *MockCargoRepository) {
				m.On("List", ctx).Return([]*domaincargo.Cargo{}, nil)
			},
			wantErr:   false,
			wantCount: 0,
		},
		{
			name: "successful list with items",
			setupMock: func(m *MockCargoRepository) {
				cargos := []*domaincargo.Cargo{
					{
						ID:       uuid.New(),
						Name:     "Cargo 1",
						Weight:   100.0,
						Status:   domaincargo.CargoStatusPending,
						VesselID: vesselID,
					},
					{
						ID:       uuid.New(),
						Name:     "Cargo 2",
						Weight:   200.0,
						Status:   domaincargo.CargoStatusInTransit,
						VesselID: vesselID,
					},
				}
				m.On("List", ctx).Return(cargos, nil)
			},
			wantErr:   false,
			wantCount: 2,
		},
		{
			name: "repository error",
			setupMock: func(m *MockCargoRepository) {
				m.On("List", ctx).Return(nil, errors.New("database connection failed"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockCargoRepository)
			tt.setupMock(mockRepo)

			uc := NewListCargosUseCase(mockRepo)
			cargos, err := uc.Execute(ctx)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, cargos)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, cargos)
				assert.Equal(t, tt.wantCount, len(cargos))
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

// TestListCargosByVesselUseCase tests the ListCargosByVesselUseCase
func TestListCargosByVesselUseCase(t *testing.T) {
	ctx := context.Background()
	vesselID := uuid.New()

	tests := []struct {
		name      string
		vesselID  uuid.UUID
		setupMock func(*MockCargoRepository)
		wantErr   bool
		wantCount int
	}{
		{
			name:     "successful list by vessel",
			vesselID: vesselID,
			setupMock: func(m *MockCargoRepository) {
				cargos := []*domaincargo.Cargo{
					{
						ID:       uuid.New(),
						Name:     "Cargo A",
						Weight:   150.0,
						Status:   domaincargo.CargoStatusPending,
						VesselID: vesselID,
					},
				}
				m.On("ListByVesselID", ctx, vesselID).Return(cargos, nil)
			},
			wantErr:   false,
			wantCount: 1,
		},
		{
			name:     "vessel has no cargo",
			vesselID: vesselID,
			setupMock: func(m *MockCargoRepository) {
				m.On("ListByVesselID", ctx, vesselID).Return([]*domaincargo.Cargo{}, nil)
			},
			wantErr:   false,
			wantCount: 0,
		},
		{
			name:     "repository error",
			vesselID: vesselID,
			setupMock: func(m *MockCargoRepository) {
				m.On("ListByVesselID", ctx, vesselID).Return(nil, errors.New("query failed"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockCargoRepository)
			tt.setupMock(mockRepo)

			uc := NewListCargosByVesselUseCase(mockRepo)
			cargos, err := uc.Execute(ctx, tt.vesselID)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, cargos)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, cargos)
				assert.Equal(t, tt.wantCount, len(cargos))
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

// TestUpdateCargoStatusUseCase tests the UpdateCargoStatusUseCase
func TestUpdateCargoStatusUseCase(t *testing.T) {
	ctx := context.Background()
	cargoID := uuid.New()
	vesselID := uuid.New()

	tests := []struct {
		name       string
		cargoID    uuid.UUID
		newStatus  domaincargo.CargoStatus
		setupMock  func(*MockCargoRepository)
		wantErr    bool
		wantStatus domaincargo.CargoStatus
	}{
		{
			name:      "successful status update",
			cargoID:   cargoID,
			newStatus: domaincargo.CargoStatusInTransit,
			setupMock: func(m *MockCargoRepository) {
				currentCargo := &domaincargo.Cargo{
					ID:       cargoID,
					Name:     "Container XYZ",
					Weight:   200.0,
					Status:   domaincargo.CargoStatusPending,
					VesselID: vesselID,
				}
				m.On("GetByID", ctx, cargoID).Return(currentCargo, nil)
				m.On("UpdateStatus", ctx, cargoID, domaincargo.CargoStatusInTransit).Return(&domaincargo.Cargo{
					ID:       cargoID,
					Name:     "Container XYZ",
					Weight:   200.0,
					Status:   domaincargo.CargoStatusInTransit,
					VesselID: vesselID,
				}, nil)
			},
			wantErr:    false,
			wantStatus: domaincargo.CargoStatusInTransit,
		},
		{
			name:      "invalid status",
			cargoID:   cargoID,
			newStatus: domaincargo.CargoStatus("invalid"),
			setupMock: func(m *MockCargoRepository) {},
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockCargoRepository)
			tt.setupMock(mockRepo)

			uc := NewUpdateCargoStatusUseCase(mockRepo, nil, nil, nil)
			cargo, err := uc.Execute(ctx, tt.cargoID, tt.newStatus)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, cargo)
				assert.Equal(t, tt.wantStatus, cargo.Status)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
