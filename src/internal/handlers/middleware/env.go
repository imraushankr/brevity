package middleware

import (
	"github.com/gin-gonic/gin"
)

func EnvMiddleware(env string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Set environment in context if needed
		c.Set("env", env)

		// Set Gin mode based on environment
		if env == "production" {
			gin.SetMode(gin.ReleaseMode)
		} else {
			gin.SetMode(gin.DebugMode)
		}

		c.Next()
	}
}
