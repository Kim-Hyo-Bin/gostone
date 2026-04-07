// Package app composes the gostone service (HTTP, storage, configuration).
package app

import (
	"fmt"
	"log"
	"os"

	"github.com/Kim-Hyo-Bin/gostone/internal/config"
	"github.com/Kim-Hyo-Bin/gostone/internal/db"
	"github.com/Kim-Hyo-Bin/gostone/internal/httpserver"
	"github.com/gin-gonic/gin"
)

// Run starts the gostone HTTP server using the merged configuration (file + env overrides).
func Run(cfg *config.Config) error {
	gdb, err := db.Open(cfg.Database.Connection)
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}

	if os.Getenv("GIN_MODE") == "" {
		if cfg.Default.Debug {
			gin.SetMode(gin.DebugMode)
		} else {
			gin.SetMode(gin.ReleaseMode)
		}
	}

	r := gin.New()
	r.Use(gin.Recovery(), gin.Logger())
	httpserver.Register(r, gdb)

	addr := cfg.Service.Listen
	log.Printf("gostone listening on %s", addr)
	if err := r.Run(addr); err != nil {
		return fmt.Errorf("http server: %w", err)
	}
	return nil
}
