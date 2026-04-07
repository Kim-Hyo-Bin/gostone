// Package httpserver registers HTTP routes for the gostone API.
package httpserver

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Register mounts routes on the given engine.
func Register(r *gin.Engine, gdb *gorm.DB) {
	r.GET("/health", func(c *gin.Context) {
		sqlDB, err := gdb.DB()
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "db_unavailable"})
			return
		}
		if err := sqlDB.Ping(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "db_ping_failed"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
}
