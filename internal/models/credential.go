package models

// Credential is a generic stored credential (non-EC2 types).
type Credential struct {
	ID        string `gorm:"primaryKey;size:64"`
	UserID    string `gorm:"not null;index;size:64"`
	ProjectID string `gorm:"index;size:64"`
	Type      string `gorm:"not null;index;size:128"`
	Blob      string `gorm:"not null;type:text"`
}
