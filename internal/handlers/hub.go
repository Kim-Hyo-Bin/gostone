package handlers

import (
	"github.com/Kim-Hyo-Bin/gostone/internal/policy"
	"github.com/Kim-Hyo-Bin/gostone/internal/token"
	"gorm.io/gorm"
)

// Hub carries dependencies shared by HTTP handlers (database, config, etc.).
type Hub struct {
	DB     *gorm.DB
	Tokens *token.JWT
	Policy *policy.Policy
}
