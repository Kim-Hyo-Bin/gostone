package models

// Endpoint is a service endpoint in the catalog (public/internal/admin).
type Endpoint struct {
	ID        string `gorm:"primaryKey;size:64"`
	ServiceID string `gorm:"not null;index;size:64"`
	RegionID  string `gorm:"not null;index;size:64"`
	Interface string `gorm:"not null;size:32"` // public, internal, admin
	URL       string `gorm:"not null;type:text"`
	Enabled   bool   `gorm:"default:true"`
}
