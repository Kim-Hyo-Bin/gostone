package models

import "time"

// AuthToken is a persisted opaque token (Keystone UUID token style).
type AuthToken struct {
	ID        string `gorm:"primaryKey;size:64"`
	UserID    string `gorm:"not null;index;size:64"`
	DomainID  string `gorm:"not null;size:64"`
	ProjectID string `gorm:"not null;size:64"`
	RolesJSON string `gorm:"type:text;not null"` // JSON array of role names
	IssuedAt  time.Time
	ExpiresAt time.Time `gorm:"not null;index"`
	RevokedAt *time.Time
}
