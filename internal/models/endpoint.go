package models

// Endpoint is a service endpoint in the catalog (public/internal/admin).
type Endpoint struct {
	ID        string `gorm:"primaryKey;type:text"`
	ServiceID string `gorm:"not null;index"`
	RegionID  string `gorm:"not null;index"`
	Interface string `gorm:"not null"` // public, internal, admin
	URL       string `gorm:"not null;type:text"`
	Enabled   bool   `gorm:"default:true"`
}
