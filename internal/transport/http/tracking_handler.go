package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/alex-necsoiu/deus-logistics-api/internal/domain/tracking"
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

// AddTrackingEntry handles POST /api/v1/cargoes/:id/tracking
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
		status := mapTrackingErrorToStatus(err.Error())
		c.JSON(status, response.ErrorResponse{
			Error: response.ErrorDetail{
				Code:      mapTrackingErrorToCode(err.Error()),
				Message:   err.Error(),
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

// GetTrackingHistory handles GET /api/v1/cargoes/:id/tracking
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
		c.JSON(http.StatusInternalServerError, response.ErrorResponse{
			Error: response.ErrorDetail{
				Code:      response.CodeInternalError,
				Message:   response.MsgFailedTrackingHistory,
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

// mapTrackingErrorToStatus maps tracking domain errors to HTTP status codes.
func mapTrackingErrorToStatus(errMsg string) int {
	switch errMsg {
	case tracking.ErrCargoNotFound.Error():
		return http.StatusNotFound
	case tracking.ErrInvalidEntry.Error():
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}

// mapTrackingErrorToCode maps tracking domain errors to error codes.
func mapTrackingErrorToCode(errMsg string) string {
	switch errMsg {
	case tracking.ErrCargoNotFound.Error():
		return response.CodeCargoNotFound
	case tracking.ErrInvalidEntry.Error():
		return response.CodeInvalidEntry
	default:
		return response.CodeInternalError
	}
}
