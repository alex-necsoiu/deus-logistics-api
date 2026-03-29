package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog"

	appcargo "github.com/alex-necsoiu/deus-logistics-api/internal/application/cargo"
	"github.com/alex-necsoiu/deus-logistics-api/internal/domain/cargo"
	httperrors "github.com/alex-necsoiu/deus-logistics-api/internal/errors"
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

// CreateCargo godoc
// @Summary Create a new cargo
// @Description Create a new cargo shipment with initial status
// @Tags cargo
// @Accept json
// @Produce json
// @Param request body CreateCargoRequest true "Cargo creation payload"
// @Success 201 {object} response.SuccessResponse{data=CargoResponse} "Cargo created successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid request or validation failed"
// @Failure 409 {object} response.ErrorResponse "Vessel not found"
// @Router /cargoes [post]
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
		status := httperrors.MapErrorToHTTPStatus(err)
		c.JSON(status, response.ErrorResponse{
			Error: response.ErrorDetail{
				Code:      string(httperrors.MapErrorToErrorCode(err)),
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

// GetCargo godoc
// @Summary Get cargo by ID
// @Description Retrieve a specific cargo shipment by its UUID
// @Tags cargo
// @Accept json
// @Produce json
// @Param id path string true "Cargo ID (UUID)" format(uuid)
// @Success 200 {object} response.SuccessResponse{data=CargoResponse} "Cargo found"
// @Failure 400 {object} response.ErrorResponse "Invalid cargo ID format"
// @Failure 404 {object} response.ErrorResponse "Cargo not found"
// @Router /cargoes/{id} [get]
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
		status := httperrors.MapErrorToHTTPStatus(err)
		c.JSON(status, response.ErrorResponse{
			Error: response.ErrorDetail{
				Code:      string(httperrors.MapErrorToErrorCode(err)),
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

// ListCargoes godoc
// @Summary List all cargoes
// @Description Retrieve all cargo shipments in the system
// @Tags cargo
// @Accept json
// @Produce json
// @Success 200 {object} response.SuccessResponse{data=[]CargoResponse} "List of cargoes"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /cargoes [get]
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

// UpdateCargoStatus godoc
// @Summary Update cargo status
// @Description Update the status of a cargo shipment (pending, in_transit, or delivered)
// @Tags cargo
// @Accept json
// @Produce json
// @Param id path string true "Cargo ID (UUID)" format(uuid)
// @Param request body UpdateCargoStatusRequest true "Status update payload"
// @Success 200 {object} response.SuccessResponse{data=CargoResponse} "Status updated successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid request or validation failed"
// @Failure 404 {object} response.ErrorResponse "Cargo not found"
// @Router /cargoes/{id}/status [patch]
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
		status := httperrors.MapErrorToHTTPStatus(err)
		c.JSON(status, response.ErrorResponse{
			Error: response.ErrorDetail{
				Code:      string(httperrors.MapErrorToErrorCode(err)),
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
