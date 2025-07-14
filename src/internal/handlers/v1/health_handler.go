package v1

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/imraushankr/brevity/server/src/configs"
)

type HealthHandler struct {
	cfg *configs.Config
}

func NewHealthHandler(cfg *configs.Config) *HealthHandler {
	return &HealthHandler{cfg: cfg}
}

func (h *HealthHandler) GetHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
		"system": gin.H{
			"version":     h.cfg.App.Version,
			"environment": h.cfg.App.Environment,
			"timestamp":   time.Now().UTC().Format(time.RFC3339),
		},
	})
}

func (h *HealthHandler) GetSystemInfo(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"app": gin.H{
			"name":        h.cfg.App.Name,
			"version":     h.cfg.App.Version,
			"environment": h.cfg.App.Environment,
			"debug":       h.cfg.App.Debug,
		},
		"server": gin.H{
			"host":             h.cfg.Server.Host,
			"port":             h.cfg.Server.Port,
			"read_timeout":     h.cfg.Server.ReadTimeout.String(),
			"write_timeout":    h.cfg.Server.WriteTimeout.String(),
			"shutdown_timeout": h.cfg.Server.ShutdownTimeout.String(),
		},
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

func (h *HealthHandler) GetConfig(c *gin.Context) {
	c.JSON(http.StatusOK, h.cfg)
}