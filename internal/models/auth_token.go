package models

import "time"

// AuthToken is a persisted opaque token (Keystone UUID token style).
type AuthToken struct {
	ID        string `gorm:"primaryKey;type:text"`
	UserID    string `gorm:"not null;index"`
	DomainID  string `gorm:"not null"`
	ProjectID string `gorm:"not null"`
	RolesJSON string `gorm:"type:text;not null"` // JSON array of role names
	IssuedAt  time.Time
	ExpiresAt time.Time `gorm:"not null;index"`
	RevokedAt *time.Time
}
