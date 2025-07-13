// package app

// import (
// 	"net/http"

// 	"github.com/gin-gonic/gin"
// 	"github.com/imraushankr/brevity/server/src/configs"
// 	"github.com/imraushankr/brevity/server/src/internal/pkg/database"
// )

// func SetupRouter(cfg *configs.Config, db *database.DB) (*gin.Engine, error) {
// 	// Set Gin mode based on environment
// 	if cfg.App.Environment == "production" {
// 		gin.SetMode(gin.ReleaseMode)
// 	} else {
// 		gin.SetMode(gin.DebugMode)
// 	}

// 	router := gin.New()

// 	// Middleware
// 	router.Use(gin.Logger())
// 	router.Use(gin.Recovery())

// 	// Health check endpoint
// 	router.GET("/health", func(c *gin.Context) {
// 		c.JSON(http.StatusOK, gin.H{
// 			"status": "ok",
// 		})
// 	})

// 	// Add your routes here
// 	// Example:
// 	// api := router.Group("/api")
// 	// {
// 	//     api.GET("/users", handlers.GetUsers(db))
// 	//     api.POST("/users", handlers.CreateUser(db))
// 	// }

// 	return router, nil
// }


package app

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/imraushankr/brevity/server/src/configs"
	"github.com/imraushankr/brevity/server/src/internal/pkg/database"
)

func SetupRouter(cfg *configs.Config, db *database.DB) (*gin.Engine, error) {
	// Set Gin mode based on environment
	if cfg.App.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	router := gin.New()

	// Middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Health check endpoint with detailed information
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"app": gin.H{
				"name":        cfg.App.Name,
				"version":     cfg.App.Version,
				"environment": cfg.App.Environment,
				"debug":       cfg.App.Debug,
			},
			"server": gin.H{
				"host": cfg.Server.Host,
				"port": cfg.Server.Port,
			},
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"uptime":    time.Since(time.Now()).String(), // You might want to track actual uptime
		})
	})

	// Server info endpoint
	router.GET("/info", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"app": gin.H{
				"name":        cfg.App.Name,
				"version":     cfg.App.Version,
				"environment": cfg.App.Environment,
				"debug":       cfg.App.Debug,
			},
			"server": gin.H{
				"host":             cfg.Server.Host,
				"port":             cfg.Server.Port,
				"read_timeout":     cfg.Server.ReadTimeout.String(),
				"write_timeout":    cfg.Server.WriteTimeout.String(),
				"shutdown_timeout": cfg.Server.ShutdownTimeout.String(),
			},
			"database": gin.H{
				"path":         cfg.Database.SQLite.Path,
				"journal_mode": cfg.Database.SQLite.JournalMode,
				"foreign_keys": cfg.Database.SQLite.ForeignKeys,
			},
			"cors": gin.H{
				"enabled": cfg.CORS.Enabled,
				"origins": cfg.CORS.AllowOrigins,
				"methods": cfg.CORS.AllowMethods,
			},
			"rate_limit": gin.H{
				"enabled":  cfg.RateLimit.Enabled,
				"requests": cfg.RateLimit.Requests,
				"window":   cfg.RateLimit.Window,
			},
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
	})

	// Root endpoint
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Welcome to " + cfg.App.Name,
			"version": cfg.App.Version,
			"status":  "running",
			"endpoints": gin.H{
				"health": "/health",
				"info":   "/info",
			},
		})
	})

	// Add your routes here
	// Example:
	// api := router.Group("/api")
	// {
	//     api.GET("/users", handlers.GetUsers(db))
	//     api.POST("/users", handlers.CreateUser(db))
	// }

	return router, nil
}