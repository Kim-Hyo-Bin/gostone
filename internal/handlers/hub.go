package handlers

import "gorm.io/gorm"

// Hub carries dependencies shared by HTTP handlers (database, config, etc.).
type Hub struct {
	DB *gorm.DB
}
