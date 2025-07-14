package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/imraushankr/brevity/server/src/internal/pkg/logger"
)

// RecoveryMiddleware creates a recovery middleware with comprehensive logging
func RecoveryMiddleware(log logger.Logger) gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		// Collect request details for logging
		fields := []logger.Field{
			logger.String("path", c.Request.URL.Path),
			logger.String("method", c.Request.Method),
			logger.String("client_ip", c.ClientIP()),
			logger.String("user_agent", c.Request.UserAgent()),
		}

		// Handle different types of panic recoveries
		switch err := recovered.(type) {
		case string:
			fields = append(fields, logger.String("error", err))
			log.Error("panic recovered: string error", fields...)
		case error:
			fields = append(fields, logger.ErrorField(err))
			log.Error("panic recovered: error type", fields...)
		default:
			fields = append(fields, logger.Any("recovered", recovered))
			log.Error("panic recovered: unknown type", fields...)
		}

		// Respond with JSON error
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal server error",
			"message": "Something went wrong. Please try again later.",
		})
	})
}