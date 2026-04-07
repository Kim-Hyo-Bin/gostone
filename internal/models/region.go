package models

// Region is an OpenStack region (catalog / scope).
type Region struct {
	ID          string `gorm:"primaryKey;type:text"`
	Description string `gorm:"type:text"`
	ParentID    string `gorm:"type:text"`
	Extra       string `gorm:"type:text"` // JSON blob for future
}
