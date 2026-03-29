package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/alex-necsoiu/deus-logistics-api/internal/domain/vessel"
	httperrors "github.com/alex-necsoiu/deus-logistics-api/internal/errors"
	"github.com/alex-necsoiu/deus-logistics-api/pkg/response"
)

// VesselHandler provides HTTP handlers for vessel operations.
type VesselHandler struct {
	service vessel.Service
}

// NewVesselHandler creates a new vessel handler.
func NewVesselHandler(service vessel.Service) *VesselHandler {
	return &VesselHandler{service: service}
}

// CreateVessel handles POST /api/v1/vessels
func (h *VesselHandler) CreateVessel(c *gin.Context) {
	ctx := c.Request.Context()
	logger := zerolog.Ctx(ctx)

	var req CreateVesselRequest
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

	input := vessel.CreateVesselInput{
		Name:            req.Name,
		Capacity:        req.Capacity,
		CurrentLocation: req.CurrentLocation,
	}

	result, err := h.service.CreateVessel(ctx, input)
	if err != nil {
		status := httperrors.MapErrorToHTTPStatus(err)
		c.JSON(status, response.ErrorResponse{
			Error: response.ErrorDetail{
				Code:      string(httperrors.MapErrorToErrorCode(err)),
				Message:   httperrors.MapErrorToErrorMessage(err),
				RequestID: c.GetString(response.CtxRequestID),
			},
		})
		return
	}

	c.JSON(http.StatusCreated, response.SuccessResponse{
		Data: VesselResponse{
			ID:              result.ID,
			Name:            result.Name,
			Capacity:        result.Capacity,
			CurrentLocation: result.CurrentLocation,
			CreatedAt:       result.CreatedAt,
			UpdatedAt:       result.UpdatedAt,
		},
		Meta: response.Meta{
			RequestID: c.GetString(response.CtxRequestID),
		},
	})
}

// GetVessel handles GET /api/v1/vessels/:id
func (h *VesselHandler) GetVessel(c *gin.Context) {
	ctx := c.Request.Context()

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: response.ErrorDetail{
				Code:      response.CodeInvalidInput,
				Message:   response.MsgInvalidVesselID,
				RequestID: c.GetString(response.CtxRequestID),
			},
		})
		return
	}

	result, err := h.service.GetVessel(ctx, id)
	if err != nil {
		status := httperrors.MapErrorToHTTPStatus(err)
		c.JSON(status, response.ErrorResponse{
			Error: response.ErrorDetail{
				Code:      string(httperrors.MapErrorToErrorCode(err)),
				Message:   httperrors.MapErrorToErrorMessage(err),
				RequestID: c.GetString(response.CtxRequestID),
			},
		})
		return
	}

	c.JSON(http.StatusOK, response.SuccessResponse{
		Data: VesselResponse{
			ID:              result.ID,
			Name:            result.Name,
			Capacity:        result.Capacity,
			CurrentLocation: result.CurrentLocation,
			CreatedAt:       result.CreatedAt,
			UpdatedAt:       result.UpdatedAt,
		},
		Meta: response.Meta{
			RequestID: c.GetString(response.CtxRequestID),
		},
	})
}

// ListVessels handles GET /api/v1/vessels
func (h *VesselHandler) ListVessels(c *gin.Context) {
	ctx := c.Request.Context()

	result, err := h.service.ListVessels(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error: response.ErrorDetail{
				Code:      response.CodeInternalError,
				Message:   response.MsgFailedListVessels,
				RequestID: c.GetString(response.CtxRequestID),
			},
		})
		return
	}

	responses := make([]VesselResponse, len(result))
	for i, v := range result {
		responses[i] = VesselResponse{
			ID:              v.ID,
			Name:            v.Name,
			Capacity:        v.Capacity,
			CurrentLocation: v.CurrentLocation,
			CreatedAt:       v.CreatedAt,
			UpdatedAt:       v.UpdatedAt,
		}
	}

	c.JSON(http.StatusOK, response.SuccessResponse{
		Data: responses,
		Meta: response.Meta{
			RequestID: c.GetString(response.CtxRequestID),
		},
	})
}

// UpdateVesselLocation handles PATCH /api/v1/vessels/:id/location
func (h *VesselHandler) UpdateVesselLocation(c *gin.Context) {
	ctx := c.Request.Context()
	logger := zerolog.Ctx(ctx)

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.ErrorResponse{
			Error: response.ErrorDetail{
				Code:      response.CodeInvalidInput,
				Message:   response.MsgInvalidVesselID,
				RequestID: c.GetString(response.CtxRequestID),
			},
		})
		return
	}

	var req UpdateVesselLocationRequest
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

	result, err := h.service.UpdateVesselLocation(ctx, id, req.CurrentLocation)
	if err != nil {
		status := httperrors.MapErrorToHTTPStatus(err)
		c.JSON(status, response.ErrorResponse{
			Error: response.ErrorDetail{
				Code:      string(httperrors.MapErrorToErrorCode(err)),
				Message:   httperrors.MapErrorToErrorMessage(err),
				RequestID: c.GetString(response.CtxRequestID),
			},
		})
		return
	}

	c.JSON(http.StatusOK, response.SuccessResponse{
		Data: VesselResponse{
			ID:              result.ID,
			Name:            result.Name,
			Capacity:        result.Capacity,
			CurrentLocation: result.CurrentLocation,
			CreatedAt:       result.CreatedAt,
			UpdatedAt:       result.UpdatedAt,
		},
		Meta: response.Meta{
			RequestID: c.GetString(response.CtxRequestID),
		},
	})
}
