package http

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog"

	appcargo "github.com/alex-necsoiu/deus-logistics-api/internal/application/cargo"
	"github.com/alex-necsoiu/deus-logistics-api/internal/domain/cargo"
	"github.com/alex-necsoiu/deus-logistics-api/pkg/response"
)

// CargoHandler provides HTTP handlers for cargo operations.
type CargoHandler struct {
	app *appcargo.CargoApplicationManager
}

// NewCargoHandler creates a new cargo handler with application layer use cases.
func NewCargoHandler(app *appcargo.CargoApplicationManager) *CargoHandler {
	return &CargoHandler{app: app}
}

// CreateCargo handles POST /api/v1/cargoes
func (h *CargoHandler) CreateCargo(c *gin.Context) {
	ctx := c.Request.Context()
	logger := zerolog.Ctx(ctx)

	var req CreateCargoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error().Err(err).Msg("invalid request body")
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: response.ErrorDetail{
				Code:      response.CodeInvalidInput,
				Message:   response.MsgInvalidRequestBody,
				RequestID: c.GetString(response.CtxRequestID),
			},
		})
		return
	}

	// Validate business rules
	if err := req.Validate(); err != nil {
		logger.Error().Err(err).Msg("validation failed")
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: response.ErrorDetail{
				Code:      response.CodeInvalidInput,
				Message:   err.Error(),
				RequestID: c.GetString(response.CtxRequestID),
			},
		})
		return
	}

	input := cargo.CreateCargoInput{
		Name:        req.Name,
		Description: req.Description,
		Weight:      req.Weight,
		VesselID:    req.VesselID,
	}

	// Execute use case: orchestration and persistence
	result, err := h.app.CreateCargo.Execute(ctx, input)
	if err != nil {
		status := mapErrorToStatus(err)
		c.JSON(status, response.ErrorResponse{
			Error: response.ErrorDetail{
				Code:      mapErrorToCode(err),
				Message:   err.Error(),
				RequestID: c.GetString(response.CtxRequestID),
			},
		})
		return
	}

	c.JSON(http.StatusCreated, response.SuccessResponse{
		Data: CargoResponse{
			ID:          result.ID,
			Name:        result.Name,
			Description: result.Description,
			Weight:      result.Weight,
			Status:      result.Status.String(),
			VesselID:    result.VesselID,
			CreatedAt:   result.CreatedAt,
			UpdatedAt:   result.UpdatedAt,
		},
		Meta: response.Meta{
			RequestID: c.GetString(response.CtxRequestID),
		},
	})
}

// GetCargo handles GET /api/v1/cargoes/:id
func (h *CargoHandler) GetCargo(c *gin.Context) {
	ctx := c.Request.Context()

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: response.ErrorDetail{
				Code:      response.CodeInvalidInput,
				Message:   response.MsgInvalidCargoID,
				RequestID: c.GetString(response.CtxRequestID),
			},
		})
		return
	}

	// Execute use case: retrieve cargo
	result, err := h.app.GetCargo.Execute(ctx, id)
	if err != nil {
		status := mapErrorToStatus(err)
		c.JSON(status, response.ErrorResponse{
			Error: response.ErrorDetail{
				Code:      mapErrorToCode(err),
				Message:   err.Error(),
				RequestID: c.GetString(response.CtxRequestID),
			},
		})
		return
	}

	c.JSON(http.StatusOK, response.SuccessResponse{
		Data: CargoResponse{
			ID:          result.ID,
			Name:        result.Name,
			Description: result.Description,
			Weight:      result.Weight,
			Status:      result.Status.String(),
			VesselID:    result.VesselID,
			CreatedAt:   result.CreatedAt,
			UpdatedAt:   result.UpdatedAt,
		},
		Meta: response.Meta{
			RequestID: c.GetString(response.CtxRequestID),
		},
	})
}

// ListCargoes handles GET /api/v1/cargoes
func (h *CargoHandler) ListCargoes(c *gin.Context) {
	ctx := c.Request.Context()

	// Execute use case: list all cargos
	result, err := h.app.ListCargos.Execute(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error: response.ErrorDetail{
				Code:      response.CodeInternalError,
				Message:   response.MsgFailedListCargoes,
				RequestID: c.GetString(response.CtxRequestID),
			},
		})
		return
	}

	responses := make([]CargoResponse, len(result))
	for i, c := range result {
		responses[i] = CargoResponse{
			ID:          c.ID,
			Name:        c.Name,
			Description: c.Description,
			Weight:      c.Weight,
			Status:      c.Status.String(),
			VesselID:    c.VesselID,
			CreatedAt:   c.CreatedAt,
			UpdatedAt:   c.UpdatedAt,
		}
	}

	c.JSON(http.StatusOK, response.SuccessResponse{
		Data: responses,
		Meta: response.Meta{
			RequestID: c.GetString(response.CtxRequestID),
		},
	})
}

// UpdateCargoStatus handles PATCH /api/v1/cargoes/:id/status
func (h *CargoHandler) UpdateCargoStatus(c *gin.Context) {
	ctx := c.Request.Context()
	logger := zerolog.Ctx(ctx)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: response.ErrorDetail{
				Code:      response.CodeInvalidInput,
				Message:   response.MsgInvalidCargoID,
				RequestID: c.GetString(response.CtxRequestID),
			},
		})
		return
	}

	var req UpdateCargoStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error().Err(err).Msg("invalid request body")
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: response.ErrorDetail{
				Code:      response.CodeInvalidInput,
				Message:   response.MsgInvalidRequestBody,
				RequestID: c.GetString(response.CtxRequestID),
			},
		})
		return
	}

	// Validate request parameters
	if err := req.Validate(); err != nil {
		logger.Error().Err(err).Msg("validation failed")
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: response.ErrorDetail{
				Code:      response.CodeInvalidInput,
				Message:   err.Error(),
				RequestID: c.GetString(response.CtxRequestID),
			},
		})
		return
	}

	newStatus := cargo.CargoStatus(req.Status)
	// Execute use case: orchestrates domain validation, persistence, tracking, and events
	result, err := h.app.UpdateStatus.Execute(ctx, id, newStatus)
	if err != nil {
		status := mapErrorToStatus(err)
		c.JSON(status, response.ErrorResponse{
			Error: response.ErrorDetail{
				Code:      mapErrorToCode(err),
				Message:   err.Error(),
				RequestID: c.GetString(response.CtxRequestID),
			},
		})
		return
	}

	c.JSON(http.StatusOK, response.SuccessResponse{
		Data: CargoResponse{
			ID:          result.ID,
			Name:        result.Name,
			Description: result.Description,
			Weight:      result.Weight,
			Status:      result.Status.String(),
			VesselID:    result.VesselID,
			CreatedAt:   result.CreatedAt,
			UpdatedAt:   result.UpdatedAt,
		},
		Meta: response.Meta{
			RequestID: c.GetString(response.CtxRequestID),
		},
	})
}

// mapErrorToStatus maps domain errors to HTTP status codes using error comparison.
func mapErrorToStatus(err error) int {
	switch {
	case errors.Is(err, cargo.ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, cargo.ErrInvalidInput), errors.Is(err, cargo.ErrInvalidStatus):
		return http.StatusBadRequest
	case errors.Is(err, cargo.ErrInvalidTransition):
		// 422 Unprocessable Entity: Valid request but business logic doesn't allow this action
		return http.StatusUnprocessableEntity
	default:
		return http.StatusInternalServerError
	}
}

// mapErrorToCode maps domain errors to error codes using error comparison.
func mapErrorToCode(err error) string {
	switch {
	case errors.Is(err, cargo.ErrNotFound):
		return response.CodeCargoNotFound
	case errors.Is(err, cargo.ErrInvalidInput):
		return response.CodeInvalidInput
	case errors.Is(err, cargo.ErrInvalidStatus), errors.Is(err, cargo.ErrInvalidTransition):
		return response.CodeInvalidStatus
	default:
		return response.CodeInternalError
	}
}
