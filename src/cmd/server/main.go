package main

import (
	"os"

	"github.com/imraushankr/brevity/server/src/configs"
	"github.com/imraushankr/brevity/server/src/internal/app"
	"github.com/imraushankr/brevity/server/src/internal/pkg/logger"
	"go.uber.org/zap"
)

func main() {
	// Initialize basic logger
	logger.InitBasic()
	defer logger.Sync()

	// Load configuration
	cfg, err := configs.LoadConfig(configs.GetConfigPath())
	if err != nil {
		logger.Fatal("Failed to load config", zap.Error(err), zap.String("hint", "Ensure app.yaml exists in configs/ directory"))
		os.Exit(1)
	}

	// Initialize full logger
	logger.Init(&cfg.Logger)
	defer logger.Sync()

	// Create server
	server, err := app.NewServer(cfg)
	if err != nil {
		logger.Fatal("Failed to create server", zap.Error(err))
	}

	if err := server.Run(); err != nil {
		logger.Fatal("Server exited with error", zap.Error(err))
	}
}