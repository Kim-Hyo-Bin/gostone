// Package app composes the gostone service (HTTP, storage, configuration).
package app

import (
	"fmt"
	"log"
	"os"

	"github.com/Kim-Hyo-Bin/gostone/internal/db"
	"github.com/Kim-Hyo-Bin/gostone/internal/httpserver"
	"github.com/gin-gonic/gin"
)

// Run starts the gostone HTTP server and backing database connection.
func Run() error {
	dsn := os.Getenv("GOSTONE_SQLITE_DSN")
	if dsn == "" {
		dsn = "file::memory:?cache=shared"
	}

	gdb, err := db.Open(dsn)
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}

	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.Use(gin.Recovery(), gin.Logger())
	httpserver.Register(r, gdb)

	addr := os.Getenv("GOSTONE_HTTP_ADDR")
	if addr == "" {
		addr = ":8080"
	}
	log.Printf("gostone listening on %s", addr)
	if err := r.Run(addr); err != nil {
		return fmt.Errorf("http server: %w", err)
	}
	return nil
}
