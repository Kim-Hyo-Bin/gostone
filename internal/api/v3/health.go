package v3

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// RegisterHealth exposes GET /health (liveness: process only) and GET /ready (readiness: DB ping).
func RegisterHealth(r *gin.Engine, h *Hub) {
	r.GET("/health", h.getHealthLiveness)
	r.GET("/ready", h.getHealthReadiness)
}

func (h *Hub) getHealthLiveness(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *Hub) getHealthReadiness(c *gin.Context) {
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
