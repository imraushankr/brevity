package routes

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/imraushankr/brevity/server/src/configs"
	"github.com/imraushankr/brevity/server/src/internal/handlers/middleware"
	handlersV1 "github.com/imraushankr/brevity/server/src/internal/handlers/v1"
	"github.com/imraushankr/brevity/server/src/internal/pkg/auth"
	"github.com/imraushankr/brevity/server/src/internal/pkg/database"
	"github.com/imraushankr/brevity/server/src/internal/pkg/email"
	"github.com/imraushankr/brevity/server/src/internal/pkg/logger"
	"github.com/imraushankr/brevity/server/src/internal/pkg/storage"
	"github.com/imraushankr/brevity/server/src/internal/repository"
	routesV1 "github.com/imraushankr/brevity/server/src/internal/routes/v1"
	"github.com/imraushankr/brevity/server/src/internal/services"
)

func SetupRoutes(router *gin.Engine, cfg *configs.Config, db *database.DB, authService *auth.Auth, log logger.Logger) (*gin.Engine, error) {
	// Global middleware
	router.Use(
		gin.Recovery(),
		middleware.RequestLogger(log),
		middleware.CORS(),
		middleware.RateLimiter(100, 10),
		middleware.PrometheusMetricsMiddleware(),
	)

	// Initialize storage service
	storageService, err := storage.NewStorage(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize storage: %w", err)
	}

	// Initialize services
	userSvc, err := initUserService(cfg, db, authService, storageService, log)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize user service: %w", err)
	}

	// Initialize handlers
	healthHandler := handlersV1.NewHealthHandler(cfg)
	userHandler := handlersV1.NewUserHandler(userSvc)

	// API routes
	api := router.Group("/api")
	{
		// Version 1 routes
		v1Group := api.Group("/v1", APIVersion("v1"))
		{
			routesV1.RegisterAuthRoutes(v1Group, userHandler, authService, cfg)
			routesV1.RegisterUserRoutes(v1Group, userHandler, authService, cfg)
			routesV1.RegisterSystemRoutes(v1Group, healthHandler)
		}

		// Add future version groups here (v2, etc.)
	}

	return router, nil
}

func initUserService(
	cfg *configs.Config,
	db *database.DB,
	authService *auth.Auth,
	storageService storage.Storage,
	log logger.Logger,
) (services.UserService, error) {
	emailService := email.NewEmailService(&cfg.Email, log)
	userRepo := repository.NewUserRepository(db.DB)

	userSvc := services.NewUserService(userRepo, authService, emailService, cfg, storageService)

	return userSvc, nil
}