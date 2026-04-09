package v3

import (
	"net/http"

	"github.com/Kim-Hyo-Bin/gostone/internal/buildinfo"
	"github.com/gin-gonic/gin"
)

// RegisterHealth exposes GET /health (liveness: process only) and GET /ready (readiness: DB ping).
func RegisterHealth(r *gin.Engine, h *Hub) {
	r.GET("/health", h.getHealthLiveness)
	r.GET("/ready", h.getHealthReadiness)
}

func (h *Hub) getHealthLiveness(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"service": "gostone",
		"version": buildinfo.Version,
		"commit":  buildinfo.Commit,
	})
}

func (h *Hub) getHealthReadiness(c *gin.Context) {
	checks := gin.H{}
	allOK := true

	sqlDB, err := h.DB.DB()
	if err != nil {
		checks["database"] = gin.H{"status": "error", "detail": err.Error()}
		allOK = false
	} else if err := sqlDB.Ping(); err != nil {
		checks["database"] = gin.H{"status": "error", "detail": err.Error()}
		allOK = false
	} else {
		checks["database"] = gin.H{"status": "ok"}
	}

	status := http.StatusOK
	body := gin.H{
		"status":  "ok",
		"service": "gostone",
		"version": buildinfo.Version,
		"commit":  buildinfo.Commit,
		"checks":  checks,
	}
	if !allOK {
		status = http.StatusServiceUnavailable
		body["status"] = "unavailable"
	}
	c.JSON(status, body)
}
