package v3

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// RegisterHealth exposes GET /health (liveness; separate from Keystone /healthcheck).
func RegisterHealth(r *gin.Engine, h *Hub) {
	r.GET("/health", h.getHealth)
}

func (h *Hub) getHealth(c *gin.Context) {
	sqlDB, err := h.DB.DB()
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "db_unavailable"})
		return
	}
	if err := sqlDB.Ping(); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"status": "db_ping_failed"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
