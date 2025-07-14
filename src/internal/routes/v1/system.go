package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/imraushankr/brevity/server/src/internal/handlers/middleware"
	"github.com/imraushankr/brevity/server/src/internal/handlers/v1"
)

func RegisterSystemRoutes(r *gin.RouterGroup, handler *v1.HealthHandler) {
	systemGroup := r.Group("/system")
	{
		// Health endpoints
		systemGroup.GET("/health", handler.GetHealth)
		systemGroup.GET("/status", handler.GetSystemInfo)

		// Configuration (only in non-production)
		if gin.Mode() != gin.ReleaseMode {
			systemGroup.GET("/config", handler.GetConfig)
		}

		// Metrics endpoints
		systemGroup.GET("/metrics", middleware.PrometheusHandler())
		systemGroup.GET("/metrics-default", middleware.PrometheusHandler())
	}
}