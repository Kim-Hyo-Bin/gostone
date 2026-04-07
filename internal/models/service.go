package models

// Service is a catalog service entry (e.g. identity, compute).
type Service struct {
	ID          string `gorm:"primaryKey;type:text"`
	Type        string `gorm:"not null;index"`
	Name        string `gorm:"not null"`
	Description string `gorm:"type:text"`
	Enabled     bool   `gorm:"default:true"`
	Extra       string `gorm:"type:text"`
}
