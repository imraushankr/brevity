// package main

// import (
// 	"os"
// 	"github.com/imraushankr/brevity/server/src/configs"
// 	"github.com/imraushankr/brevity/server/src/internal/app"
// 	"github.com/imraushankr/brevity/server/src/internal/pkg/logger"
// )

// func main() {
// 	// Initialize basic logger with default configuration
// 	log, err := logger.Init(logger.Config{
// 		Level:  "info",
// 		Format: "console",
// 	})
// 	if err != nil {
// 		panic("Failed to initialize basic logger: " + err.Error())
// 	}
// 	defer logger.Sync()

// 	// Load configuration
// 	cfg, err := configs.LoadConfig(configs.GetConfigPath())
// 	if err != nil {
// 		log.Fatal("Failed to load config", 
// 			logger.ErrorField(err), 
// 			logger.String("hint", "Ensure app.yaml exists in configs/ directory"))
// 		os.Exit(1)
// 	}

// 	// Convert configs.LoggerConfig to logger.Config
// 	loggerCfg := logger.Config{
// 		Level:    cfg.Logger.Level,
// 		Format:   cfg.Logger.Format,
// 		FilePath: cfg.Logger.FilePath,
// 	}

// 	// Initialize full logger with config
// 	log, err = logger.Init(loggerCfg)
// 	if err != nil {
// 		log.Fatal("Failed to initialize logger", logger.ErrorField(err))
// 		os.Exit(1)
// 	}
// 	defer logger.Sync()

// 	// Create server
// 	server, err := app.NewServer(cfg)
// 	if err != nil {
// 		log.Fatal("Failed to create server", logger.ErrorField(err))
// 	}

// 	if err := server.Run(); err != nil {
// 		log.Fatal("Server exited with error", logger.ErrorField(err))
// 	}
// }


package main

import (
	"os"

	"github.com/imraushankr/brevity/server/src/configs"
	"github.com/imraushankr/brevity/server/src/internal/app"
	"github.com/imraushankr/brevity/server/src/internal/pkg/logger"
)

func main() {
	// Initialize basic logger with default configuration
	log, err := logger.Init(logger.Config{
		Level:  "info",
		Format: "console",
	})
	if err != nil {
		panic("Failed to initialize basic logger: " + err.Error())
	}
	defer logger.Sync()

	// Load configuration
	cfg, err := configs.LoadConfig(configs.GetConfigPath())
	if err != nil {
		log.Fatal("Failed to load config", 
			logger.ErrorField(err), 
			logger.String("hint", "Ensure app.yaml exists in configs/ directory"))
		os.Exit(1)
	}

	// Convert configs.LoggerConfig to logger.Config
	loggerCfg := logger.Config{
		Level:    cfg.Logger.Level,
		Format:   cfg.Logger.Format,
		FilePath: cfg.Logger.FilePath,
	}

	// Initialize full logger with config
	log, err = logger.Init(loggerCfg)
	if err != nil {
		log.Fatal("Failed to initialize logger", logger.ErrorField(err))
		os.Exit(1)
	}
	defer logger.Sync()

	// Create server
	server, err := app.NewServer(cfg)
	if err != nil {
		log.Fatal("Failed to create server", logger.ErrorField(err))
	}

	if err := server.Run(); err != nil {
		log.Fatal("Server exited with error", logger.ErrorField(err))
	}
}