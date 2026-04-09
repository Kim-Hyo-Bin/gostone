package models

// RegisteredLimit defines default limits for a resource (Keystone unified limits).
type RegisteredLimit struct {
	ID           string `gorm:"primaryKey;size:64"`
	ServiceID    string `gorm:"index;size:64"`
	RegionID     string `gorm:"index;size:64"`
	ResourceName string `gorm:"not null;size:255"`
	Default      int64  `gorm:"not null"`
	Description  string `gorm:"type:text"`
}

// Limit is a project-scoped override of a registered limit.
type Limit struct {
	ID                string `gorm:"primaryKey;size:64"`
	ProjectID         string `gorm:"not null;index;size:64"`
	RegisteredLimitID string `gorm:"not null;index;size:64"`
	ResourceLimit     int64  `gorm:"not null"`
}
