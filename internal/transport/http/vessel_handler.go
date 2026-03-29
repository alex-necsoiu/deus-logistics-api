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

// CreateVessel godoc
// @Summary Create a new vessel
// @Description Create a new vessel with specified capacity and location
// @Tags vessel
// @Accept json
// @Produce json
// @Param request body CreateVesselRequest true "Vessel creation payload"
// @Success 201 {object} response.SuccessResponse{data=VesselResponse} "Vessel created successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid request or validation failed"
// @Router /vessels [post]
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

// GetVessel godoc
// @Summary Get vessel by ID
// @Description Retrieve a specific vessel by its UUID
// @Tags vessel
// @Accept json
// @Produce json
// @Param id path string true "Vessel ID (UUID)" format(uuid)
// @Success 200 {object} response.SuccessResponse{data=VesselResponse} "Vessel found"
// @Failure 400 {object} response.ErrorResponse "Invalid vessel ID format"
// @Failure 404 {object} response.ErrorResponse "Vessel not found"
// @Router /vessels/{id} [get]
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

// ListVessels godoc
// @Summary List all vessels
// @Description Retrieve all vessels in the system
// @Tags vessel
// @Accept json
// @Produce json
// @Success 200 {object} response.SuccessResponse{data=[]VesselResponse} "List of vessels"
// @Failure 500 {object} response.ErrorResponse "Internal server error"
// @Router /vessels [get]
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

// UpdateVesselLocation godoc
// @Summary Update vessel location
// @Description Update the current location of a vessel
// @Tags vessel
// @Accept json
// @Produce json
// @Param id path string true "Vessel ID (UUID)" format(uuid)
// @Param request body UpdateVesselLocationRequest true "Location update payload"
// @Success 200 {object} response.SuccessResponse{data=VesselResponse} "Location updated successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid request or validation failed"
// @Failure 404 {object} response.ErrorResponse "Vessel not found"
// @Router /vessels/{id}/location [patch]
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
