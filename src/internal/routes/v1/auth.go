package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/imraushankr/brevity/server/src/configs"
	"github.com/imraushankr/brevity/server/src/internal/handlers/middleware"
	"github.com/imraushankr/brevity/server/src/internal/handlers/v1"
	"github.com/imraushankr/brevity/server/src/internal/pkg/auth"
)

func RegisterAuthRoutes(r *gin.RouterGroup, handler *v1.UserHandler, authService *auth.Auth, cfg *configs.Config) {
	authGroup := r.Group("/auth")
	{
		// Public endpoints
		authGroup.POST("/register", handler.Register)
		authGroup.POST("/login", handler.Login)
		authGroup.GET("/verify-email", handler.VerifyEmail)
		authGroup.POST("/password-reset", handler.InitiatePasswordReset)
		authGroup.POST("/password-reset/confirm", handler.CompletePasswordReset)

		// Refresh token endpoint (requires valid refresh token)
		refreshGroup := authGroup.Group("", middleware.RefreshTokenAuth(authService, cfg))
		refreshGroup.POST("/refresh", handler.RefreshToken)
	}
}