package routes

import (
	"github.com/gin-gonic/gin"
)

// APIVersion sets the API version in the context
func APIVersion(version string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("api-version", version)
		c.Header("X-API-Version", version)
		c.Next()
	}
}

// DeprecationNotice marks deprecated API versions
func DeprecationNotice(version string, sunsetDate string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-API-Deprecated", "true")
		c.Header("X-API-Sunset-Date", sunsetDate)
		c.Next()
	}
}