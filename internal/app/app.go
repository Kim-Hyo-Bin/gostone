// Package app composes the gostone service (HTTP, storage, configuration).
package app

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Kim-Hyo-Bin/gostone/internal/api/v3"
	"github.com/Kim-Hyo-Bin/gostone/internal/bootstrap"
	"github.com/Kim-Hyo-Bin/gostone/internal/conf"
	"github.com/Kim-Hyo-Bin/gostone/internal/db"
	"github.com/Kim-Hyo-Bin/gostone/internal/policy"
	"github.com/Kim-Hyo-Bin/gostone/internal/server"
	"github.com/Kim-Hyo-Bin/gostone/internal/token"
	"github.com/gin-gonic/gin"
)

// Run starts the gostone HTTP server using the merged configuration (file + env overrides).
func Run(cfg *conf.Config) error {
	if cfg.Token.Secret == "" {
		return fmt.Errorf("token signing secret is empty: set [token] secret in gostone.conf or GOSTONE_TOKEN_SECRET")
	}
	ttl := time.Duration(cfg.Token.ExpirationHours) * time.Hour
	if cfg.Token.ExpirationHours <= 0 {
		ttl = 24 * time.Hour
	}
	jwtIssuer := &token.JWT{
		Secret: []byte(cfg.Token.Secret),
		Issuer: "gostone",
		TTL:    ttl,
	}

	gdb, err := db.Open(cfg.Database.Connection)
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	if err := db.AutoMigrate(gdb); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}
	if err := bootstrap.FromEnv(gdb); err != nil {
		return fmt.Errorf("bootstrap: %w", err)
	}

	if os.Getenv("GIN_MODE") == "" {
		if cfg.Default.Debug {
			gin.SetMode(gin.DebugMode)
		} else {
			gin.SetMode(gin.ReleaseMode)
		}
	}

	hub := &v3.Hub{
		DB:     gdb,
		Tokens: jwtIssuer,
		Policy: policy.Default(),
	}

	r := gin.New()
	r.Use(gin.Recovery(), gin.Logger())
	server.Register(r, hub)

	addr := cfg.Service.Listen
	log.Printf("gostone listening on %s", addr)
	if err := r.Run(addr); err != nil {
		return fmt.Errorf("http server: %w", err)
	}
	return nil
}
