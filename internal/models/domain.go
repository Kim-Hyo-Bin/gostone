package models

// Domain is an Identity domain (Keystone domain).
type Domain struct {
	ID      string `gorm:"primaryKey;size:64"`
	Name    string `gorm:"not null;uniqueIndex:ux_domains_name;size:255"`
	Enabled bool   `gorm:"default:true"`
}
