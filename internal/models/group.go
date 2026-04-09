package models

// Group is a Keystone identity group within a domain.
type Group struct {
	ID          string `gorm:"primaryKey;size:64"`
	Name        string `gorm:"not null;index;size:255"`
	DomainID    string `gorm:"not null;index;size:64"`
	Description string `gorm:"type:text"`
}

// GroupMember links a user to a group.
type GroupMember struct {
	ID      uint   `gorm:"primaryKey"`
	GroupID string `gorm:"not null;index:idx_group_member,priority:1;uniqueIndex:idx_group_user;size:64"`
	UserID  string `gorm:"not null;index;uniqueIndex:idx_group_user;size:64"`
}
