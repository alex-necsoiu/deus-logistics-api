package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/alex-necsoiu/deus-logistics-api/internal/domain/tracking"
	httperrors "github.com/alex-necsoiu/deus-logistics-api/internal/errors"
	"github.com/alex-necsoiu/deus-logistics-api/pkg/response"
)

// TrackingHandler provides HTTP handlers for tracking operations.
type TrackingHandler struct {
	service tracking.Service
}

// NewTrackingHandler creates a new tracking handler.
func NewTrackingHandler(service tracking.Service) *TrackingHandler {
	return &TrackingHandler{service: service}
}

// AddTrackingEntry godoc
// @Summary Add a tracking entry
// @Description Add a new tracking entry for a cargo shipment
// @Tags tracking
// @Accept json
// @Produce json
// @Param id path string true "Cargo ID (UUID)" format(uuid)
// @Param request body AddTrackingRequest true "Tracking entry payload"
// @Success 201 {object} response.SuccessResponse{data=TrackingEntryResponse} "Tracking entry created successfully"
// @Failure 400 {object} response.ErrorResponse "Invalid request or validation failed"
// @Failure 404 {object} response.ErrorResponse "Cargo not found"
// @Router /cargoes/{id}/tracking [post]
func (h *TrackingHandler) AddTrackingEntry(c *gin.Context) {
	ctx := c.Request.Context()
	logger := zerolog.Ctx(ctx)

	cargoID, err := uuid.Parse(c.Param("id"))
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

	var req AddTrackingRequest
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

	input := tracking.AddTrackingInput{
		CargoID:  cargoID,
		Location: req.Location,
		Status:   req.Status,
		Note:     req.Note,
	}

	result, err := h.service.AddTrackingEntry(ctx, input)
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
		Data: TrackingEntryResponse{
			ID:        result.ID,
			CargoID:   result.CargoID,
			Location:  result.Location,
			Status:    result.Status,
			Note:      result.Note,
			Timestamp: result.Timestamp,
		},
		Meta: response.Meta{
			RequestID: c.GetString(response.CtxRequestID),
		},
	})
}

// GetTrackingHistory godoc
// @Summary Get cargo tracking history
// @Description Retrieve the complete tracking history for a cargo shipment
// @Tags tracking
// @Accept json
// @Produce json
// @Param id path string true "Cargo ID (UUID)" format(uuid)
// @Success 200 {object} response.SuccessResponse{data=[]TrackingEntryResponse} "Tracking history retrieved"
// @Failure 400 {object} response.ErrorResponse "Invalid cargo ID format"
// @Failure 404 {object} response.ErrorResponse "Cargo not found"
// @Router /cargoes/{id}/tracking [get]
func (h *TrackingHandler) GetTrackingHistory(c *gin.Context) {
	ctx := c.Request.Context()

	cargoID, err := uuid.Parse(c.Param("id"))
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

	result, err := h.service.GetTrackingHistory(ctx, cargoID)
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

	responses := make([]TrackingEntryResponse, len(result))
	for i, t := range result {
		responses[i] = TrackingEntryResponse{
			ID:        t.ID,
			CargoID:   t.CargoID,
			Location:  t.Location,
			Status:    t.Status,
			Note:      t.Note,
			Timestamp: t.Timestamp,
		}
	}

	c.JSON(http.StatusOK, response.SuccessResponse{
		Data: responses,
		Meta: response.Meta{
			RequestID: c.GetString(response.CtxRequestID),
		},
	})
}
