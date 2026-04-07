package db

import (
	"github.com/Kim-Hyo-Bin/gostone/internal/models"
	"gorm.io/gorm"
)

// AutoMigrate creates or updates schema for core Identity tables.
func AutoMigrate(gdb *gorm.DB) error {
	return gdb.AutoMigrate(
		&models.Domain{},
		&models.Project{},
		&models.User{},
		&models.Role{},
		&models.UserProjectRole{},
	)
}
