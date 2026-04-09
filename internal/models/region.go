package models

// Region is an OpenStack region (catalog / scope).
type Region struct {
	ID          string `gorm:"primaryKey;size:64"`
	Description string `gorm:"type:text"`
	ParentID    string `gorm:"size:64"`
	Extra       string `gorm:"type:text"` // JSON blob for future
}
