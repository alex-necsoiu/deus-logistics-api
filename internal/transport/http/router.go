package http

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	appcargo "github.com/alex-necsoiu/deus-logistics-api/internal/application/cargo"
	"github.com/alex-necsoiu/deus-logistics-api/internal/domain/tracking"
	"github.com/alex-necsoiu/deus-logistics-api/internal/domain/vessel"
	"github.com/alex-necsoiu/deus-logistics-api/internal/health"
	"github.com/alex-necsoiu/deus-logistics-api/pkg/response"
)

// Router registers all HTTP routes and middleware.
func Router(
	engine *gin.Engine,
	db *pgxpool.Pool,
	cargoApp *appcargo.CargoApplicationManager,
	vesselSvc vessel.Service,
	trackingSvc tracking.Service,
) {
	// Middleware
	engine.Use(requestIDMiddleware())
	engine.Use(loggingMiddleware())
	engine.Use(recoveryMiddleware())

	// Initialize health reporter
	healthReporter := health.NewReporter(db)

	// Health and readiness endpoints
	engine.GET("/health", healthCheckHandler(healthReporter))
	engine.GET("/ready", readinessCheckHandler(healthReporter))

	// API routes
	api := engine.Group("/api/v1")

	// Cargo routes - using application layer use cases
	cargoHandler := NewCargoHandler(cargoApp)
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

// requestIDMiddleware injects a unique request ID into the context and headers.
func requestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader(response.HeaderRequestID)
		if requestID == "" {
			requestID = uuid.New().String()
		}
		c.Set(response.CtxRequestID, requestID)
		c.Header(response.HeaderRequestID, requestID)

		// Add request_id to context for downstream middleware and handlers
		ctx := c.Request.Context()
		ctx = context.WithValue(ctx, response.CtxRequestID, requestID)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

// loggingMiddleware logs incoming requests and responses with structured JSON format.
// Attaches enriched logger with request_id to context so downstream handlers can use
// zerolog.Ctx(ctx) and produce real log output.
func loggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetString(response.CtxRequestID)

		// Attach enriched logger to context so zerolog.Ctx(ctx) works downstream
		logger := log.Logger.With().Str("request_id", requestID).Logger()
		ctx := logger.WithContext(c.Request.Context())
		c.Request = c.Request.WithContext(ctx)

		c.Next()

		zerolog.Ctx(c.Request.Context()).Info().
			Str("method", c.Request.Method).
			Str("path", c.Request.URL.Path).
			Int("status", c.Writer.Status()).
			Int("content_length", c.Writer.Size()).
			Msg("http request completed")
	}
}

// recoveryMiddleware recovers from panics and returns 500 error with proper logging.
func recoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				requestID := c.GetString(response.CtxRequestID)
				log := zerolog.Ctx(c.Request.Context())
				log.Error().
					Interface("panic", err).
					Str("request_id", requestID).
					Str("path", c.Request.URL.Path).
					Msg("http panic recovered")

				c.JSON(http.StatusInternalServerError, response.ErrorResponse{
					Error: response.ErrorDetail{
						Code:      response.CodeInternalError,
						Message:   response.MsgInternalServerError,
						RequestID: requestID,
					},
				})
			}
		}()
		c.Next()
	}
}

// healthCheckHandler handles GET /health (liveness probe)
func healthCheckHandler(reporter *health.Reporter) gin.HandlerFunc {
	return func(c *gin.Context) {
		check := reporter.CheckLiveness(c.Request.Context())
		if check.Status != health.StatusHealthy {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "unhealthy",
				"check":  check,
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
			"check":  check,
		})
	}
}

// readinessCheckHandler handles GET /ready (readiness probe)
func readinessCheckHandler(reporter *health.Reporter) gin.HandlerFunc {
	return func(c *gin.Context) {
		result := reporter.CheckReadiness(c.Request.Context())
		statusCode := http.StatusOK
		if !result.Ready {
			statusCode = http.StatusServiceUnavailable
		}
		c.JSON(statusCode, result)
	}
}
