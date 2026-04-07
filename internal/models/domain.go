package models

// Domain is an Identity domain (Keystone domain).
type Domain struct {
	ID      string `gorm:"primaryKey;type:text"`
	Name    string `gorm:"not null;uniqueIndex:ux_domains_name"`
	Enabled bool   `gorm:"default:true"`
}
