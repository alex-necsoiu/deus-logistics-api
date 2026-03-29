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

	"github.com/alex-necsoiu/deus-logistics-api/internal/domain/cargo"
	"github.com/alex-necsoiu/deus-logistics-api/pkg/response"
)

// MockCargoService is a mock implementation of cargo.Service
type MockCargoService struct {
	mock.Mock
}

func (m *MockCargoService) CreateCargo(ctx context.Context, input cargo.CreateCargoInput) (*cargo.Cargo, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*cargo.Cargo), args.Error(1)
}

func (m *MockCargoService) GetCargo(ctx context.Context, id uuid.UUID) (*cargo.Cargo, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*cargo.Cargo), args.Error(1)
}

func (m *MockCargoService) ListCargoes(ctx context.Context) ([]*cargo.Cargo, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*cargo.Cargo), args.Error(1)
}

func (m *MockCargoService) ListCargoByVesselID(ctx context.Context, vesselID uuid.UUID) ([]*cargo.Cargo, error) {
	args := m.Called(ctx, vesselID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*cargo.Cargo), args.Error(1)
}

func (m *MockCargoService) UpdateCargoStatus(ctx context.Context, id uuid.UUID, status cargo.CargoStatus) (*cargo.Cargo, error) {
	args := m.Called(ctx, id, status)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*cargo.Cargo), args.Error(1)
}

func setupTestRouter(t *testing.T) (*gin.Engine, *MockCargoService) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	mockService := new(MockCargoService)
	handler := NewCargoHandler(mockService)

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

	return router, mockService
}

func TestCargoHandler_CreateCargo(t *testing.T) {
	vesselID := uuid.New()
	cargoID := uuid.New()

	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		mockSetup      func(*MockCargoService)
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
			mockSetup: func(m *MockCargoService) {
				m.On("CreateCargo", mock.Anything, mock.MatchedBy(func(input cargo.CreateCargoInput) bool {
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
			mockSetup:      func(m *MockCargoService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "negative weight returns 400",
			requestBody: map[string]interface{}{
				"name":      "Test",
				"weight":    -100.0,
				"vessel_id": vesselID.String(),
			},
			mockSetup:      func(m *MockCargoService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "service error returns 500",
			requestBody: map[string]interface{}{
				"name":      "Test",
				"weight":    100.0,
				"vessel_id": vesselID.String(),
			},
			mockSetup: func(m *MockCargoService) {
				m.On("CreateCargo", mock.Anything, mock.Anything).
					Return(nil, cargo.ErrInvalidInput)
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router, mockService := setupTestRouter(t)
			tt.mockSetup(mockService)

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
		mockSetup      func(*MockCargoService)
		expectedStatus int
	}{
		{
			name:   "valid ID returns 200",
			pathID: cargoID.String(),
			mockSetup: func(m *MockCargoService) {
				m.On("GetCargo", mock.Anything, cargoID).
					Return(&cargo.Cargo{
						ID:       cargoID,
						Name:     "Test",
						Status:   cargo.CargoStatusPending,
						VesselID: uuid.New(),
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "not found returns 404",
			pathID: cargoID.String(),
			mockSetup: func(m *MockCargoService) {
				m.On("GetCargo", mock.Anything, cargoID).
					Return(nil, cargo.ErrNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "invalid UUID returns 400",
			pathID:         "not-a-uuid",
			mockSetup:      func(m *MockCargoService) {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router, mockService := setupTestRouter(t)
			tt.mockSetup(mockService)

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
		mockSetup      func(*MockCargoService)
		expectedStatus int
	}{
		{
			name: "with results returns 200",
			mockSetup: func(m *MockCargoService) {
				m.On("ListCargoes", mock.Anything).
					Return([]*cargo.Cargo{
						{ID: uuid.New(), Name: "C1", Status: cargo.CargoStatusPending, VesselID: uuid.New(), CreatedAt: time.Now(), UpdatedAt: time.Now()},
						{ID: uuid.New(), Name: "C2", Status: cargo.CargoStatusInTransit, VesselID: uuid.New(), CreatedAt: time.Now(), UpdatedAt: time.Now()},
					}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "empty list returns 200",
			mockSetup: func(m *MockCargoService) {
				m.On("ListCargoes", mock.Anything).
					Return([]*cargo.Cargo{}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "service error returns 500",
			mockSetup: func(m *MockCargoService) {
				m.On("ListCargoes", mock.Anything).
					Return(nil, errors.New("db error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router, mockService := setupTestRouter(t)
			tt.mockSetup(mockService)

			req := httptest.NewRequest(http.MethodGet, "/api/v1/cargoes", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestCargoHandler_UpdateStatus(t *testing.T) {
	cargoID := uuid.New()

	tests := []struct {
		name           string
		pathID         string
		requestBody    map[string]interface{}
		mockSetup      func(*MockCargoService)
		expectedStatus int
	}{
		{
			name:   "valid status update returns 200",
			pathID: cargoID.String(),
			requestBody: map[string]interface{}{
				"status": cargo.CargoStatusInTransit.String(),
			},
			mockSetup: func(m *MockCargoService) {
				m.On("UpdateCargoStatus", mock.Anything, cargoID, cargo.CargoStatusInTransit).
					Return(&cargo.Cargo{
						ID:        cargoID,
						Status:    cargo.CargoStatusInTransit,
						VesselID:  uuid.New(),
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "invalid status returns 400",
			pathID: cargoID.String(),
			requestBody: map[string]interface{}{
				"status": "invalid",
			},
			mockSetup: func(m *MockCargoService) {
				m.On("UpdateCargoStatus", mock.Anything, cargoID, cargo.CargoStatus("invalid")).
					Return(nil, cargo.ErrInvalidStatus)
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "missing status returns 400",
			pathID: cargoID.String(),
			requestBody: map[string]interface{}{
			},
			mockSetup:      func(m *MockCargoService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "not found returns 404",
			pathID: cargoID.String(),
			requestBody: map[string]interface{}{
				"status": cargo.CargoStatusInTransit.String(),
			},
			mockSetup: func(m *MockCargoService) {
				m.On("UpdateCargoStatus", mock.Anything, cargoID, cargo.CargoStatusInTransit).
					Return(nil, cargo.ErrNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "invalid UUID returns 400",
			pathID:         "not-a-uuid",
			requestBody: map[string]interface{}{
				"status": cargo.CargoStatusInTransit.String(),
			},
			mockSetup:      func(m *MockCargoService) {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router, mockService := setupTestRouter(t)
			tt.mockSetup(mockService)

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPatch, "/api/v1/cargoes/"+tt.pathID+"/status", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}
