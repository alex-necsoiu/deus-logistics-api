package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/alex-necsoiu/deus-logistics-api/internal/domain/cargo"
	"github.com/alex-necsoiu/deus-logistics-api/pkg/response"
)

// CargoHandler provides HTTP handlers for cargo operations.
type CargoHandler struct {
	service cargo.Service
}

// NewCargoHandler creates a new cargo handler.
func NewCargoHandler(service cargo.Service) *CargoHandler {
	return &CargoHandler{service: service}
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

	result, err := h.service.CreateCargo(ctx, input)
	if err != nil {
		status := mapErrorToStatus(err.Error())
		c.JSON(status, response.ErrorResponse{
			Error: response.ErrorDetail{
				Code:      mapErrorToCode(err.Error()),
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

	result, err := h.service.GetCargo(ctx, id)
	if err != nil {
		status := mapErrorToStatus(err.Error())
		c.JSON(status, response.ErrorResponse{
			Error: response.ErrorDetail{
				Code:      mapErrorToCode(err.Error()),
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

	result, err := h.service.ListCargoes(ctx)
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

	status := cargo.CargoStatus(req.Status)
	result, err := h.service.UpdateCargoStatus(ctx, id, status)
	if err != nil {
		status := mapErrorToStatus(err.Error())
		c.JSON(status, response.ErrorResponse{
			Error: response.ErrorDetail{
				Code:      mapErrorToCode(err.Error()),
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

// mapErrorToStatus maps domain errors to HTTP status codes.
func mapErrorToStatus(errMsg string) int {
	switch errMsg {
	case cargo.ErrNotFound.Error():
		return http.StatusNotFound
	case cargo.ErrInvalidInput.Error(), cargo.ErrInvalidStatus.Error():
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}

// mapErrorToCode maps domain errors to error codes.
func mapErrorToCode(errMsg string) string {
	switch errMsg {
	case cargo.ErrNotFound.Error():
		return response.CodeCargoNotFound
	case cargo.ErrInvalidInput.Error():
		return response.CodeInvalidInput
	case cargo.ErrInvalidStatus.Error():
		return response.CodeInvalidStatus
	default:
		return response.CodeInternalError
	}
}
