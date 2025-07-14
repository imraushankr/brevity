package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/imraushankr/brevity/server/src/configs"
	"github.com/imraushankr/brevity/server/src/internal/pkg/database"
	"github.com/imraushankr/brevity/server/src/internal/pkg/logger"
	"go.uber.org/zap"
)

type Server struct {
	httpServer *http.Server
	db         *database.DB
	cfg        *configs.Config
	router     *gin.Engine
}

func NewServer(cfg *configs.Config) (*Server, error) {
	// Initialize database
	db, err := database.ConnectDB(&cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Verify database connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.Ping(ctx); err != nil {
		return nil, fmt.Errorf("database ping failed: %w", err)
	}

	// Initialize logger
	log := logger.Get()

	// Initialize router
	router, err := SetupRouter(cfg, db, log)
	if err != nil {
		return nil, fmt.Errorf("failed to setup router: %w", err)
	}

	return &Server{
		httpServer: &http.Server{
			Addr:         cfg.Server.Host + ":" + cfg.Server.Port,
			Handler:      router,
			ReadTimeout:  cfg.Server.ReadTimeout,
			WriteTimeout: cfg.Server.WriteTimeout,
			IdleTimeout:  30 * time.Second,
		},
		db:     db,
		cfg:    cfg,
		router: router,
	}, nil
}

func (s *Server) displayServerInfo() {
	boxWidth := 60
	line := strings.Repeat("=", boxWidth)
	
	fmt.Printf("\n%s\n", line)
	fmt.Printf("%-*s\n", boxWidth, fmt.Sprintf("  🚀 %s Server Starting", s.cfg.App.Name))
	fmt.Printf("%s\n", line)
	
	fmt.Printf("%-20s: %s\n", "  App Name", s.cfg.App.Name)
	fmt.Printf("%-20s: %s\n", "  Version", s.cfg.App.Version)
	fmt.Printf("%-20s: %s\n", "  Environment", strings.ToUpper(s.cfg.App.Environment))
	fmt.Printf("%-20s: %s\n", "  Host", s.cfg.Server.Host)
	fmt.Printf("%-20s: %s\n", "  Port", s.cfg.Server.Port)
	
	protocol := "http"
	if s.cfg.JWT.SecureCookie {
		protocol = "https"
	}
	
	baseURL := fmt.Sprintf("%s://%s:%s", protocol, s.cfg.Server.Host, s.cfg.Server.Port)
	localURL := fmt.Sprintf("%s://localhost:%s", protocol, s.cfg.Server.Port)
	
	fmt.Printf("%-20s: %s\n", "  Base URL", baseURL)
	fmt.Printf("%-20s: %s\n", "  Local URL", localURL)
	fmt.Printf("%-20s: %s/health\n", "  Health Check", localURL)
	
	fmt.Printf("%-20s: %v\n", "  Debug Mode", s.cfg.App.Debug)
	fmt.Printf("%-20s: %v\n", "  Read Timeout", s.cfg.Server.ReadTimeout)
	fmt.Printf("%-20s: %v\n", "  Write Timeout", s.cfg.Server.WriteTimeout)
	
	fmt.Printf("%s\n", line)
	fmt.Printf("  📡 Server is ready to accept connections\n")
	fmt.Printf("  🔗 Visit: %s\n", localURL)
	fmt.Printf("  ❤️  Health: %s/health\n", localURL)
	fmt.Printf("  🛑 Press Ctrl+C to stop\n")
	fmt.Printf("%s\n\n", line)
}

func (s *Server) Start() error {
	s.displayServerInfo()
	
	zap.L().Info("Starting server",
		zap.String("app_name", s.cfg.App.Name),
		zap.String("version", s.cfg.App.Version),
		zap.String("address", s.httpServer.Addr),
		zap.String("environment", s.cfg.App.Environment),
		zap.Bool("debug", s.cfg.App.Debug),
		zap.Duration("read_timeout", s.cfg.Server.ReadTimeout),
		zap.Duration("write_timeout", s.cfg.Server.WriteTimeout),
	)

	serverErr := make(chan error, 1)
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErr <- err
		}
	}()

	select {
	case err := <-serverErr:
		return fmt.Errorf("server failed to start: %w", err)
	case <-time.After(100 * time.Millisecond):
		return nil
	}
}

func (s *Server) Stop(ctx context.Context) error {
	fmt.Println("\n🛑 Shutting down server gracefully...")
	
	zap.L().Info("Shutting down server...")
	shutdownErr := make(chan error, 1)
	go func() {
		shutdownErr <- s.httpServer.Shutdown(ctx)
	}()

	select {
	case err := <-shutdownErr:
		if err != nil {
			return fmt.Errorf("server shutdown failed: %w", err)
		}
	case <-ctx.Done():
		return fmt.Errorf("server shutdown timed out: %w", ctx.Err())
	}

	if err := s.db.Close(); err != nil {
		zap.L().Error("Failed to close database", zap.Error(err))
		return fmt.Errorf("database shutdown failed: %w", err)
	}

	fmt.Println("✅ Server shutdown complete")
	zap.L().Info("Server shutdown complete")
	return nil
}

func (s *Server) Run() error {
	if err := s.Start(); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), s.cfg.Server.ShutdownTimeout)
	defer cancel()

	return s.Stop(ctx)
}