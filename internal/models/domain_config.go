package models

// DomainConfig stores domain-scoped configuration options.
type DomainConfig struct {
	ID       uint   `gorm:"primaryKey"`
	DomainID string `gorm:"not null;index:idx_dom_cfg,priority:1;uniqueIndex:idx_dom_cfg_opt;size:64"`
	Group    string `gorm:"column:cfg_group;not null;uniqueIndex:idx_dom_cfg_opt;size:128"`
	Option   string `gorm:"column:cfg_option;not null;uniqueIndex:idx_dom_cfg_opt;size:128"`
	Value    string `gorm:"type:text"`
}
