// Package db opens GORM database connections (SQLite today; more drivers later).
package db

import (
	"fmt"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// Open returns a GORM DB using the given SQLite DSN (e.g. file::memory:?cache=shared).
func Open(sqliteDSN string) (*gorm.DB, error) {
	gdb, err := gorm.Open(sqlite.Open(sqliteDSN), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("gorm open: %w", err)
	}
	return gdb, nil
}
