package models

import "time"

// RevokeEvent is an OS-REVOKE audit event (minimal storage).
type RevokeEvent struct {
	ID        uint      `gorm:"primaryKey"`
	AuditID   string    `gorm:"index;size:64"`
	DomainID  string    `gorm:"index;size:64"`
	Reason    string    `gorm:"type:text"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}
