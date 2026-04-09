package db

import (
	"time"

	"gorm.io/gorm"
)

// ConfigureConnPool applies sql.DB pool settings when values are > 0.
func ConfigureConnPool(gdb *gorm.DB, maxOpen, maxIdle int, connMaxLifetime time.Duration) error {
	if gdb == nil {
		return nil
	}
	sqlDB, err := gdb.DB()
	if err != nil {
		return err
	}
	if maxOpen > 0 {
		sqlDB.SetMaxOpenConns(maxOpen)
	}
	if maxIdle > 0 {
		sqlDB.SetMaxIdleConns(maxIdle)
	}
	if connMaxLifetime > 0 {
		sqlDB.SetConnMaxLifetime(connMaxLifetime)
	}
	return nil
}
