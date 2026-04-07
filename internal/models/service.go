package models

// Service is a catalog service entry (e.g. identity, compute).
type Service struct {
	ID          string `gorm:"primaryKey;size:64"`
	Type        string `gorm:"not null;index;size:128"`
	Name        string `gorm:"not null;size:255"`
	Description string `gorm:"type:text"`
	Enabled     bool   `gorm:"default:true"`
	Extra       string `gorm:"type:text"`
}
