// Package httpserver registers HTTP routes for the gostone API.
package httpserver

import (
	"github.com/Kim-Hyo-Bin/gostone/internal/handlers"
	"github.com/gin-gonic/gin"
)

// Register mounts routes on the given engine.
func Register(r *gin.Engine, h *handlers.Hub) {
	handlers.Register(r, h)
}
