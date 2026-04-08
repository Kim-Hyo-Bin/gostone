package app

import (
	"fmt"
	"strings"

	"github.com/Kim-Hyo-Bin/gostone/internal/bootstrap"
	"github.com/Kim-Hyo-Bin/gostone/internal/conf"
	"github.com/Kim-Hyo-Bin/gostone/internal/db"
)

// Bootstrap runs database migrations, seeds the initial admin (empty user table only),
// and ensures a minimal identity catalog — similar to keystone-manage db_sync + bootstrap.
func Bootstrap(cfg *conf.Config, o bootstrap.Options) error {
	if cfg == nil {
		return fmt.Errorf("nil config")
	}
	gdb, err := db.Open(cfg.Database.Connection)
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	if err := db.AutoMigrate(gdb); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}
	if err := bootstrap.RunAdmin(gdb, o); err != nil {
		return err
	}
	publicURL := cfg.Service.PublicURL
	catalogRegion := strings.TrimSpace(o.RegionID)
	if catalogRegion == "" {
		catalogRegion = strings.TrimSpace(cfg.Service.RegionID)
	}
	if err := bootstrap.EnsureIdentityCatalog(gdb, publicURL,
		strings.TrimSpace(cfg.Service.AdminURL),
		strings.TrimSpace(cfg.Service.InternalURL),
		catalogRegion,
	); err != nil {
		return fmt.Errorf("catalog: %w", err)
	}
	return nil
}
