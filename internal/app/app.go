// Package app composes the gostone service (HTTP, storage, configuration).
package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/Kim-Hyo-Bin/gostone/internal/api/discovery"
	"github.com/Kim-Hyo-Bin/gostone/internal/api/v3"
	"github.com/Kim-Hyo-Bin/gostone/internal/bootstrap"
	"github.com/Kim-Hyo-Bin/gostone/internal/conf"
	"github.com/Kim-Hyo-Bin/gostone/internal/db"
	"github.com/Kim-Hyo-Bin/gostone/internal/listenctx"
	"github.com/Kim-Hyo-Bin/gostone/internal/policy"
	"github.com/Kim-Hyo-Bin/gostone/internal/server"
	"github.com/Kim-Hyo-Bin/gostone/internal/token"
	"github.com/gin-gonic/gin"
	"golang.org/x/sync/errgroup"
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
	discovery.ConfigureDiscovery(discovery.DocConfig{
		VersionID: strings.TrimSpace(cfg.Discovery.APIVersionID),
		Updated:   strings.TrimSpace(cfg.Discovery.Updated),
		Status:    strings.TrimSpace(cfg.Discovery.Status),
	})
	discovery.SetTrustForwardedHeaders(cfg.Service.TrustForwardedHeaders)
	if err := bootstrap.EnsureIdentityCatalog(gdb,
		strings.TrimSpace(cfg.Service.PublicURL),
		strings.TrimSpace(cfg.Service.AdminURL),
		strings.TrimSpace(cfg.Service.InternalURL),
		strings.TrimSpace(cfg.Service.RegionID),
	); err != nil {
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

	bindings, err := conf.ListenBindings(&cfg.Service)
	if err != nil {
		return err
	}
	engine := server.NewEngine(hub, server.EngineOptions{
		EnforceAdminOnly:  cfg.Service.EnforceAdminOnlyRoutes,
		AdminOnlyPrefixes: conf.ParseCommaList(cfg.Service.AdminOnlyPathPrefixes),
	})

	shutdownSec := cfg.Service.ShutdownTimeoutSeconds
	if shutdownSec <= 0 {
		shutdownSec = 15
	}

	servers := make([]*http.Server, len(bindings))
	var g errgroup.Group
	for i := range bindings {
		b := bindings[i]
		h := listenctx.WrapHandler(b.Name, engine)
		srv := &http.Server{Addr: b.Addr, Handler: h}
		servers[i] = srv
		g.Go(func() error {
			log.Printf("gostone listening (%s) %s", b.Name, b.Addr)
			err := srv.ListenAndServe()
			if err == http.ErrServerClosed {
				return nil
			}
			if err != nil {
				return fmt.Errorf("http %s %s: %w", b.Name, b.Addr, err)
			}
			return nil
		})
	}

	errCh := make(chan error, 1)
	go func() { errCh <- g.Wait() }()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	select {
	case err := <-errCh:
		return err
	case <-sigCh:
		log.Print("gostone: shutting down...")
		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(shutdownSec)*time.Second)
		defer cancel()
		for _, srv := range servers {
			if srv == nil {
				continue
			}
			if err := srv.Shutdown(ctx); err != nil {
				log.Printf("gostone: shutdown: %v", err)
			}
		}
		_ = g.Wait()
		log.Print("gostone: stopped")
		return nil
	}
}
