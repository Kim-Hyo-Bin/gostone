package handlers

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Register wires all HTTP routes (health, discovery, Identity v3).
func Register(r *gin.Engine, db *gorm.DB) {
	h := &Hub{DB: db}
	registerHealth(r, h)
	registerDiscovery(r, h)
	registerV3(r.Group("/v3"), h)
}
