package models

// Role is a named role; DomainID empty means global (Keystone-style).
type Role struct {
	ID       string `gorm:"primaryKey;size:64"`
	Name     string `gorm:"not null;uniqueIndex:ux_role_domain_name;size:255"`
	DomainID string `gorm:"uniqueIndex:ux_role_domain_name;size:64"` // optional scope
}
