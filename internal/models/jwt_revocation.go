package models

import "time"

// JWTRevocation stores revoked JWT "jti" values after DELETE /v3/auth/tokens.
type JWTRevocation struct {
	JTI       string    `gorm:"primaryKey;size:128"`
	RevokedAt time.Time `gorm:"not null"`
}
