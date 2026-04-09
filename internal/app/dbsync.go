package app

import (
	"fmt"
	"time"

	"github.com/Kim-Hyo-Bin/gostone/internal/conf"
	"github.com/Kim-Hyo-Bin/gostone/internal/db"
)

// DBSync opens the configured database and applies schema migrations (GORM AutoMigrate),
// then returns. It does not start HTTP, bootstrap the catalog, or seed admin users —
// comparable in spirit to `keystone-manage db_sync`.
func DBSync(cfg *conf.Config) error {
	if cfg == nil {
		return fmt.Errorf("nil config")
	}
	if err := conf.Validate(cfg); err != nil {
		return err
	}
	gdb, err := db.Open(cfg.Database.Connection)
	if err != nil {
		return fmt.Errorf("database: %w", err)
	}
	lifetime := time.Duration(cfg.Database.ConnMaxLifetimeSeconds) * time.Second
	if err := db.ConfigureConnPool(gdb, cfg.Database.MaxOpenConns, cfg.Database.MaxIdleConns, lifetime); err != nil {
		return fmt.Errorf("database pool: %w", err)
	}
	if err := db.AutoMigrate(gdb); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}
	return nil
}
