package v3

import (
	"github.com/Kim-Hyo-Bin/gostone/internal/policy"
	"github.com/Kim-Hyo-Bin/gostone/internal/token"
	"gorm.io/gorm"
)

// Hub carries dependencies shared by HTTP handlers (database, config, etc.).
type Hub struct {
	DB        *gorm.DB
	Tokens    *token.Manager
	Policy    *policy.Policy
	PublicURL string // optional advertised public base URL (scheme://host:port)
}
