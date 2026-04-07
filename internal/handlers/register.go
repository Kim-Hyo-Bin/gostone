package handlers

import (
	"github.com/Kim-Hyo-Bin/gostone/internal/auth"
	"github.com/gin-gonic/gin"
)

// Register wires all HTTP routes (health, discovery, Identity v3).
func Register(r *gin.Engine, h *Hub) {
	registerHealth(r, h)
	registerDiscovery(r, h)
	v3 := r.Group("/v3")
	v3.Use(auth.Middleware(h.Tokens))
	registerV3(v3, h)
}
