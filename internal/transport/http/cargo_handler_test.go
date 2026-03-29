package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	appcargo "github.com/alex-necsoiu/deus-logistics-api/internal/application/cargo"
	"github.com/alex-necsoiu/deus-logistics-api/internal/domain/cargo"
	"github.com/alex-necsoiu/deus-logistics-api/internal/domain/tracking"
	"github.com/alex-necsoiu/deus-logistics-api/pkg/response"
)

// MockCargoRepository mocks cargo repository for testing
type MockCargoRepository struct {
	mock.Mock
}

func (m *MockCargoRepository) Create(ctx context.Context, input cargo.CreateCargoInput) (*cargo.Cargo, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*cargo.Cargo), args.Error(1)
}

func (m *MockCargoRepository) GetByID(ctx context.Context, id uuid.UUID) (*cargo.Cargo, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*cargo.Cargo), args.Error(1)
}

func (m *MockCargoRepository) List(ctx context.Context) ([]*cargo.Cargo, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*cargo.Cargo), args.Error(1)
}

func (m *MockCargoRepository) ListByVesselID(ctx context.Context, vesselID uuid.UUID) ([]*cargo.Cargo, error) {
	args := m.Called(ctx, vesselID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*cargo.Cargo), args.Error(1)
}

func (m *MockCargoRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status cargo.CargoStatus) (*cargo.Cargo, error) {
	args := m.Called(ctx, id, status)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*cargo.Cargo), args.Error(1)
}

// MockEventPublisher mocks event publisher
type MockEventPublisher struct {
	mock.Mock
}

func (m *MockEventPublisher) PublishStatusChanged(ctx context.Context, event cargo.StatusChangedEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

// MockTrackingRepository mocks tracking repository
type MockTrackingRepository struct {
	mock.Mock
}

func (m *MockTrackingRepository) Append(ctx context.Context, input tracking.AddTrackingInput) (*tracking.TrackingEntry, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*tracking.TrackingEntry), args.Error(1)
}

func setupTestRouter(t *testing.T) (*gin.Engine, *MockCargoRepository) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	mockCargoRepo := new(MockCargoRepository)
	mockPublisher := new(MockEventPublisher)
	mockTrackingRepo := new(MockTrackingRepository)

	// Set up default behavior for tracking repo: allow any Create calls
	mockTrackingRepo.On("Append", mock.Anything, mock.Anything).
		Return(&tracking.TrackingEntry{}, nil)

	// Set up default behavior for publisher: allow any PublishStatusChanged calls
	mockPublisher.On("PublishStatusChanged", mock.Anything, mock.Anything).
		Return(nil)

	// Create application manager with real use cases and mocked repositories
	appManager := appcargo.NewCargoApplicationManager(mockCargoRepo, mockTrackingRepo, mockPublisher)
	handler := NewCargoHandler(appManager)

	router := gin.New()
	router.POST("/api/v1/cargoes", func(c *gin.Context) {
		c.Set(response.CtxRequestID, "test-request-id")
		c.Request = c.Request.WithContext(context.Background())
		handler.CreateCargo(c)
	})
	router.GET("/api/v1/cargoes", func(c *gin.Context) {
		c.Set(response.CtxRequestID, "test-request-id")
		c.Request = c.Request.WithContext(context.Background())
		handler.ListCargoes(c)
	})
	router.GET("/api/v1/cargoes/:id", func(c *gin.Context) {
		c.Set(response.CtxRequestID, "test-request-id")
		c.Request = c.Request.WithContext(context.Background())
		handler.GetCargo(c)
	})
	router.PATCH("/api/v1/cargoes/:id/status", func(c *gin.Context) {
		c.Set(response.CtxRequestID, "test-request-id")
		c.Request = c.Request.WithContext(context.Background())
		handler.UpdateCargoStatus(c)
	})

	return router, mockCargoRepo
}

func TestCargoHandler_CreateCargo(t *testing.T) {
	vesselID := uuid.New()
	cargoID := uuid.New()

	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		mockSetup      func(*MockCargoRepository)
		expectedStatus int
		shouldHaveData bool
	}{
		{
			name: "valid request returns 201 Created",
			requestBody: map[string]interface{}{
				"name":        "Test Cargo",
				"description": "Test Description",
				"weight":      100.0,
				"vessel_id":   vesselID.String(),
			},
			mockSetup: func(m *MockCargoRepository) {
				m.On("Create", mock.Anything, mock.MatchedBy(func(input cargo.CreateCargoInput) bool {
					return input.Name == "Test Cargo"
				})).Return(&cargo.Cargo{
					ID:          cargoID,
					Name:        "Test Cargo",
					Description: "Test Description",
					Weight:      100.0,
					Status:      cargo.CargoStatusPending,
					VesselID:    vesselID,
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}, nil)
			},
			expectedStatus: http.StatusCreated,
			shouldHaveData: true,
		},
		{
			name: "missing name returns 400",
			requestBody: map[string]interface{}{
				"weight":    100.0,
				"vessel_id": vesselID.String(),
			},
			mockSetup:      func(m *MockCargoRepository) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "negative weight returns 400",
			requestBody: map[string]interface{}{
				"name":      "Test",
				"weight":    -100.0,
				"vessel_id": vesselID.String(),
			},
			mockSetup:      func(m *MockCargoRepository) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "service error returns 400",
			requestBody: map[string]interface{}{
				"name":      "Test",
				"weight":    100.0,
				"vessel_id": vesselID.String(),
			},
			mockSetup: func(m *MockCargoRepository) {
				m.On("Create", mock.Anything, mock.Anything).
					Return(nil, cargo.ErrInvalidInput)
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router, mockRepo := setupTestRouter(t)
			tt.mockSetup(mockRepo)

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/api/v1/cargoes", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var resp map[string]interface{}
			json.Unmarshal(w.Body.Bytes(), &resp)

			if tt.shouldHaveData {
				assert.Contains(t, resp, "data")
			} else {
				assert.Contains(t, resp, "error")
			}
		})
	}
}

func TestCargoHandler_GetCargo(t *testing.T) {
	cargoID := uuid.New()

	tests := []struct {
		name           string
		pathID         string
		mockSetup      func(*MockCargoRepository)
		expectedStatus int
	}{
		{
			name:   "valid ID returns 200",
			pathID: cargoID.String(),
			mockSetup: func(m *MockCargoRepository) {
				m.On("GetByID", mock.Anything, cargoID).
					Return(&cargo.Cargo{
						ID:        cargoID,
						Name:      "Test",
						Status:    cargo.CargoStatusPending,
						VesselID:  uuid.New(),
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "not found returns 404",
			pathID: cargoID.String(),
			mockSetup: func(m *MockCargoRepository) {
				m.On("GetByID", mock.Anything, cargoID).
					Return(nil, cargo.ErrNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "invalid UUID returns 400",
			pathID:         "not-a-uuid",
			mockSetup:      func(m *MockCargoRepository) {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router, mockRepo := setupTestRouter(t)
			tt.mockSetup(mockRepo)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/cargoes/"+tt.pathID, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestCargoHandler_ListCargoes(t *testing.T) {
	tests := []struct {
		name           string
		mockSetup      func(*MockCargoRepository)
		expectedStatus int
	}{
		{
			name: "with results returns 200",
			mockSetup: func(m *MockCargoRepository) {
				m.On("List", mock.Anything).
					Return([]*cargo.Cargo{
						{ID: uuid.New(), Name: "C1", Status: cargo.CargoStatusPending, VesselID: uuid.New(), CreatedAt: time.Now(), UpdatedAt: time.Now()},
						{ID: uuid.New(), Name: "C2", Status: cargo.CargoStatusInTransit, VesselID: uuid.New(), CreatedAt: time.Now(), UpdatedAt: time.Now()},
					}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "empty list returns 200",
			mockSetup: func(m *MockCargoRepository) {
				m.On("List", mock.Anything).
					Return([]*cargo.Cargo{}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "service error returns 500",
			mockSetup: func(m *MockCargoRepository) {
				m.On("List", mock.Anything).
					Return(nil, errors.New("db error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router, mockRepo := setupTestRouter(t)
			tt.mockSetup(mockRepo)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/cargoes", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestCargoHandler_UpdateCargoStatus(t *testing.T) {
	cargoID := uuid.New()
	vesselID := uuid.New()

	tests := []struct {
		name           string
		cargoID        uuid.UUID
		requestBody    map[string]interface{}
		mockSetup      func(*MockCargoRepository)
		expectedStatus int
	}{
		{
			name:    "valid transition returns 200",
			cargoID: cargoID,
			requestBody: map[string]interface{}{
				"status": "in_transit",
			},
			mockSetup: func(m *MockCargoRepository) {
				// First call: GetByID to retrieve current cargo
				m.On("GetByID", mock.Anything, cargoID).
					Return(&cargo.Cargo{
						ID:        cargoID,
						Name:      "Test",
						Status:    cargo.CargoStatusPending,
						VesselID:  vesselID,
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					}, nil)
				// Second call: UpdateStatus to persist transition
				m.On("UpdateStatus", mock.Anything, cargoID, cargo.CargoStatus("in_transit")).
					Return(&cargo.Cargo{
						ID:        cargoID,
						Name:      "Test",
						Status:    cargo.CargoStatus("in_transit"),
						VesselID:  vesselID,
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:    "invalid transition returns 422",
			cargoID: cargoID,
			requestBody: map[string]interface{}{
				"status": "delivered",
			},
			mockSetup: func(m *MockCargoRepository) {
				// Current cargo is pending, trying to jump to delivered (invalid)
				m.On("GetByID", mock.Anything, cargoID).
					Return(&cargo.Cargo{
						ID:        cargoID,
						Name:      "Test",
						Status:    cargo.CargoStatusPending,
						VesselID:  vesselID,
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					}, nil)
			},
			expectedStatus: http.StatusUnprocessableEntity,
		},
		{
			name:    "cargo not found returns 404",
			cargoID: cargoID,
			requestBody: map[string]interface{}{
				"status": "in_transit",
			},
			mockSetup: func(m *MockCargoRepository) {
				m.On("GetByID", mock.Anything, cargoID).
					Return(nil, cargo.ErrNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router, mockRepo := setupTestRouter(t)
			tt.mockSetup(mockRepo)

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPatch, "/api/v1/cargoes/"+tt.cargoID.String()+"/status", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}
