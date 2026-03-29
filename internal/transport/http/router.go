package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/alex-necsoiu/deus-logistics-api/internal/domain/cargo"
	"github.com/alex-necsoiu/deus-logistics-api/internal/domain/tracking"
	"github.com/alex-necsoiu/deus-logistics-api/internal/domain/vessel"
	"github.com/alex-necsoiu/deus-logistics-api/pkg/response"
)

// Router registers all HTTP routes and middleware.
func Router(
	engine *gin.Engine,
	cargoSvc cargo.Service,
	vesselSvc vessel.Service,
	trackingSvc tracking.Service,
) {
	// Middleware
	engine.Use(requestIDMiddleware())
	engine.Use(loggingMiddleware())
	engine.Use(recoveryMiddleware())

	// Health and readiness endpoints
	engine.GET("/health", healthCheck)
	engine.GET("/ready", readinessCheck)

	// API routes
	api := engine.Group("/api/v1")

	// Cargo routes
	cargoHandler := NewCargoHandler(cargoSvc)
	api.POST("/cargoes", cargoHandler.CreateCargo)
	api.GET("/cargoes", cargoHandler.ListCargoes)
	api.GET("/cargoes/:id", cargoHandler.GetCargo)
	api.PATCH("/cargoes/:id/status", cargoHandler.UpdateCargoStatus)

	// Vessel routes
	vesselHandler := NewVesselHandler(vesselSvc)
	api.POST("/vessels", vesselHandler.CreateVessel)
	api.GET("/vessels", vesselHandler.ListVessels)
	api.GET("/vessels/:id", vesselHandler.GetVessel)
	api.PATCH("/vessels/:id/location", vesselHandler.UpdateVesselLocation)

	// Tracking routes
	trackingHandler := NewTrackingHandler(trackingSvc)
	api.POST("/cargoes/:id/tracking", trackingHandler.AddTrackingEntry)
	api.GET("/cargoes/:id/tracking", trackingHandler.GetTrackingHistory)
}

// requestIDMiddleware injects a unique request ID into the context.
func requestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader(response.HeaderRequestID)
		if requestID == "" {
			requestID = uuid.New().String()
		}
		c.Set(response.CtxRequestID, requestID)
		c.Header(response.HeaderRequestID, requestID)
		c.Next()
	}
}

// loggingMiddleware logs incoming requests and responses.
func loggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := zerolog.New(nil)
		ctx := logger.WithContext(c.Request.Context())
		c.Request = c.Request.WithContext(ctx)

		c.Next()

		log := zerolog.Ctx(c.Request.Context())
		log.Info().
			Str("method", c.Request.Method).
			Str("path", c.Request.URL.Path).
			Int("status", c.Writer.Status()).
			Str(response.CtxRequestID, c.GetString(response.CtxRequestID)).
			Msg("http request")
	}
}

// recoveryMiddleware recovers from panics and returns 500 error.
func recoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				zerolog.Ctx(c.Request.Context()).Error().
					Interface("panic", err).
					Msg("http panic recovered")

				c.JSON(http.StatusInternalServerError, response.ErrorResponse{
					Error: response.ErrorDetail{
						Code:      response.CodeInternalError,
						Message:   response.MsgInternalServerError,
						RequestID: c.GetString(response.CtxRequestID),
					},
				})
			}
		}()
		c.Next()
	}
}

// healthCheck handles GET /health
func healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}

// readinessCheck handles GET /ready
func readinessCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ready",
	})
}
