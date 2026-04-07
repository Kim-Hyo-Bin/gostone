package models

// ProjectTag is a single tag string on a project.
type ProjectTag struct {
	ID        uint   `gorm:"primaryKey"`
	ProjectID string `gorm:"not null;index:idx_proj_tag,priority:1;uniqueIndex:idx_proj_tag_val;size:64"`
	Value     string `gorm:"not null;size:255;uniqueIndex:idx_proj_tag_val"`
}
