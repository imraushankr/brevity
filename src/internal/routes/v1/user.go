package v1

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/imraushankr/brevity/server/src/configs"
	"github.com/imraushankr/brevity/server/src/internal/handlers/middleware"
	"github.com/imraushankr/brevity/server/src/internal/handlers/v1"
	"github.com/imraushankr/brevity/server/src/internal/pkg/auth"
)

func RegisterUserRoutes(r *gin.RouterGroup, handler *v1.UserHandler, authService *auth.Auth, cfg *configs.Config) {
	// Authenticated routes
	userGroup := r.Group("/users", middleware.AuthMiddleware(authService, &cfg.JWT))
	{
		// User profile management
		userGroup.GET("/:id", handler.GetUserProfile)
		userGroup.PUT("/:id", handler.UpdateUserProfile)
		// userGroup.DELETE("/:id", handler.DeleteUser)
		
		// Avatar management
		userGroup.POST("/:id/avatar", handler.UploadAvatar)
		
		// Admin-only routes
		adminGroup := userGroup.Group("", middleware.RoleMiddleware("admin"))
		{
			// adminGroup.GET("", handler.ListUsers)
			// adminGroup.PUT("/:id/role", handler.ChangeUserRole)
			adminGroup.GET("", func(ctx *gin.Context) {
				fmt.Println("users lists routes")
			})
			adminGroup.PUT("/:id/role", func(ctx *gin.Context) {
				fmt.Println("change the user role")
			})
		}
	}
}