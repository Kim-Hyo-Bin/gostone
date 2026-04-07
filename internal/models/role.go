package models

// Role is a named role; DomainID empty means global (Keystone-style).
type Role struct {
	ID       string `gorm:"primaryKey;type:text"`
	Name     string `gorm:"not null;uniqueIndex:ux_role_domain_name"`
	DomainID string `gorm:"uniqueIndex:ux_role_domain_name"` // optional scope
}
