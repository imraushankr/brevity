package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/imraushankr/brevity/server/src/configs"
	"github.com/imraushankr/brevity/server/src/internal/pkg/auth"
	"github.com/imraushankr/brevity/server/src/internal/models"
)

// AuthMiddleware creates a Gin middleware for JWT authentication
func AuthMiddleware(authService *auth.Auth, cfg *configs.JWTConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from header or cookie
		var tokenString string
		authHeader := c.GetHeader("Authorization")
		
		// Check for Bearer token in header
		if strings.HasPrefix(authHeader, "Bearer ") {
			tokenString = strings.TrimPrefix(authHeader, "Bearer ")
		} else {
			// Fallback to cookie if configured
			if cfg.SecureCookie {
				cookie, err := c.Cookie("access_token")
				if err != nil {
					c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
						"error": "Authorization token required",
					})
					return
				}
				tokenString = cookie
			} else {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"error": "Authorization header with Bearer token required",
				})
				return
			}
		}

		// Validate token
		claims, err := authService.VerifyAccessToken(tokenString)
		if err != nil {
			status := http.StatusUnauthorized
			if err == models.ErrExpiredToken {
				status = http.StatusForbidden
			}
			c.AbortWithStatusJSON(status, gin.H{
				"error": "Invalid token",
				"details": err.Error(),
			})
			return
		}

		// Set user context
		c.Set("user_id", claims.UserId)
		c.Set("user_role", claims.Role)

		c.Next()
	}
}

// RoleMiddleware creates a middleware to check user roles
func RoleMiddleware(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, exists := c.Get("user_role")
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "Forbidden - no role information",
			})
			return
		}

		// Check if user has any of the allowed roles
		for _, role := range allowedRoles {
			if role == userRole {
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"error": "Forbidden - insufficient permissions",
		})
	}
}

// RefreshTokenAuth creates a Gin middleware for refresh token authentication
func RefreshTokenAuth(authService *auth.Auth, cfg *configs.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from cookie (refresh tokens should only come via secure cookie)
		tokenString, err := c.Cookie("refresh_token")
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Refresh token required in cookie",
			})
			return
		}

		// Validate refresh token
		claims, err := authService.VerifyRefreshToken(tokenString)
		if err != nil {
			status := http.StatusUnauthorized
			if err == models.ErrExpiredToken {
				status = http.StatusForbidden
			}
			c.AbortWithStatusJSON(status, gin.H{
				"error": "Invalid refresh token",
				"details": err.Error(),
			})
			return
		}

		// Set user context
		c.Set("user_id", claims.UserId)
		c.Set("user_role", claims.Role)

		c.Next()
	}
}