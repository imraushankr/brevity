package app

import (
	"github.com/gin-gonic/gin"
	"github.com/imraushankr/brevity/server/src/configs"
	"github.com/imraushankr/brevity/server/src/internal/pkg/auth"
	"github.com/imraushankr/brevity/server/src/internal/pkg/database"
	"github.com/imraushankr/brevity/server/src/internal/pkg/logger"
	"github.com/imraushankr/brevity/server/src/internal/routes"
)

func SetupRouter(cfg *configs.Config, db *database.DB, log logger.Logger) (*gin.Engine, error) {
	router := gin.New()

	// Set Gin mode based on config
	if cfg.App.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	// Initialize auth service
	authService := auth.NewAuth(&cfg.JWT)

	// Setup all routes
	return routes.SetupRoutes(router, cfg, db, authService, log)
}