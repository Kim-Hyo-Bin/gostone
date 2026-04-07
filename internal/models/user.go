package models

// User is an Identity user within a domain.
type User struct {
	ID           string `gorm:"primaryKey;size:64"`
	DomainID     string `gorm:"not null;uniqueIndex:ux_user_domain_name;size:64"`
	Name         string `gorm:"not null;uniqueIndex:ux_user_domain_name;size:255"`
	Enabled      bool   `gorm:"default:true"`
	PasswordHash string `gorm:"not null"` // bcrypt, local auth only
}
