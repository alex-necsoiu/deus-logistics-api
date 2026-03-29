package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"github.com/alex-necsoiu/deus-logistics-api/pkg/response"
)

// ValidateJSONContentType middleware ensures requests with bodies have correct content type.
// POST, PUT, PATCH requests with a body must have Content-Type: application/json
// GET and DELETE requests are exempt from this check.
func ValidateJSONContentType() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only check for POST, PUT, PATCH requests
		if c.Request.Method != "GET" && c.Request.Method != "DELETE" {
			contentType := c.ContentType()
			if contentType != "application/json" && c.Request.ContentLength > 0 {
				logger := zerolog.Ctx(c.Request.Context())
				logger.Warn().Str("content_type", contentType).Msg("invalid content type")

				c.JSON(http.StatusBadRequest, response.ErrorResponse{
					Error: response.ErrorDetail{
						Code:      response.CodeInvalidInput,
						Message:   "Content-Type must be application/json",
						RequestID: c.GetString(response.CtxRequestID),
					},
				})
				c.Abort()
				return
			}
		}
		c.Next()
	}
}
