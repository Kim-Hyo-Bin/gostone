// Package app composes the gostone service (HTTP, storage, configuration).
package app

import (
	"fmt"
	"log"
	"os"
	"strings"
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
	ttl := time.Duration(cfg.Token.ExpirationHours) * time.Hour
	if cfg.Token.ExpirationHours <= 0 {
		ttl = 24 * time.Hour
	}
	prov := strings.ToLower(strings.TrimSpace(cfg.Token.Provider))
	if prov == token.ProviderJWT && cfg.Token.Secret == "" {
		return fmt.Errorf("token signing secret is empty for provider=jwt: set [token] secret or GOSTONE_TOKEN_SECRET")
	}

	gdb, err := db.Open(cfg.Database.Connection)
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	if err := db.AutoMigrate(gdb); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}
	if err := bootstrap.EnsureIdentityCatalog(gdb); err != nil {
		return fmt.Errorf("catalog bootstrap: %w", err)
	}
	if err := bootstrap.FromEnv(gdb); err != nil {
		return fmt.Errorf("bootstrap: %w", err)
	}

	authMethods := conf.ParseCommaList(cfg.Auth.Methods)
	mgr, err := token.NewManagerWithConfig(token.Config{
		DB:            gdb,
		Provider:      cfg.Token.Provider,
		TTL:           ttl,
		JWTSecret:     cfg.Token.Secret,
		FernetKeyRepo: strings.TrimSpace(cfg.FernetTokens.KeyRepository),
		AuthMethods:   authMethods,
	})
	if err != nil {
		return err
	}

	pol := policy.Default()
	if path := strings.TrimSpace(cfg.Policy.File); path != "" {
		loaded, err := policy.LoadFile(path)
		if err != nil {
			return fmt.Errorf("policy file %s: %w", path, err)
		}
		pol = loaded
	}

	if os.Getenv("GIN_MODE") == "" {
		if cfg.Default.Debug {
			gin.SetMode(gin.DebugMode)
		} else {
			gin.SetMode(gin.ReleaseMode)
		}
	}

	hub := &v3.Hub{
		DB:        gdb,
		Tokens:    mgr,
		Policy:    pol,
		PublicURL: cfg.Service.PublicURL,
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
