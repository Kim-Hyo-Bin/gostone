package models

// User is an Identity user within a domain.
type User struct {
	ID           string `gorm:"primaryKey;type:text"`
	DomainID     string `gorm:"not null;uniqueIndex:ux_user_domain_name"`
	Name         string `gorm:"not null;uniqueIndex:ux_user_domain_name"`
	Enabled      bool   `gorm:"default:true"`
	PasswordHash string `gorm:"not null"` // bcrypt, local auth only
}
