package models

// Project is an Identity project (tenant) scoped to a domain.
type Project struct {
	ID       string `gorm:"primaryKey;size:64"`
	DomainID string `gorm:"not null;uniqueIndex:ux_project_domain_name;size:64"`
	Name     string `gorm:"not null;uniqueIndex:ux_project_domain_name;size:255"`
	Enabled  bool   `gorm:"default:true"`
}
