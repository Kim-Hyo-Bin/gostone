package models

import "time"

// Trust delegates limited access between users (Keystone OS-TRUST).
type Trust struct {
	ID                string     `gorm:"primaryKey;size:64"`
	TrustorUserID     string     `gorm:"not null;index;size:64"`
	TrusteeUserID     string     `gorm:"not null;index;size:64"`
	Impersonation     bool       `gorm:"default:false"`
	ExpiresAt         *time.Time `gorm:"index"`
	AllowRedelegation bool       `gorm:"default:false"`
	ProjectID         string     `gorm:"type:text"`
	Extra             string     `gorm:"type:text"`
}

// TrustRole is a role allowed on a trust.
type TrustRole struct {
	ID      uint   `gorm:"primaryKey"`
	TrustID string `gorm:"not null;index;size:64"`
	RoleID  string `gorm:"not null;index;size:64"`
}
