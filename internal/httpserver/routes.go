// Package httpserver registers HTTP routes for the gostone API.
package httpserver

import (
	"github.com/Kim-Hyo-Bin/gostone/internal/handlers"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Register mounts routes on the given engine.
func Register(r *gin.Engine, gdb *gorm.DB) {
	handlers.Register(r, gdb)
}
