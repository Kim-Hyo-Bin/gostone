package models

// UserDomainRole assigns a role to a user on a domain.
type UserDomainRole struct {
	ID       uint   `gorm:"primaryKey"`
	UserID   string `gorm:"not null;index:idx_udr_user;size:64"`
	DomainID string `gorm:"not null;index:idx_udr_domain;size:64"`
	RoleID   string `gorm:"not null;index:idx_udr_role;size:64"`
}

// GroupProjectRole assigns a role to a group on a project.
type GroupProjectRole struct {
	ID        uint   `gorm:"primaryKey"`
	GroupID   string `gorm:"not null;index;size:64"`
	ProjectID string `gorm:"not null;index;size:64"`
	RoleID    string `gorm:"not null;index;size:64"`
}

// GroupDomainRole assigns a role to a group on a domain.
type GroupDomainRole struct {
	ID       uint   `gorm:"primaryKey"`
	GroupID  string `gorm:"not null;index;size:64"`
	DomainID string `gorm:"not null;index;size:64"`
	RoleID   string `gorm:"not null;index;size:64"`
}

// UserSystemRole assigns a role at system scope.
type UserSystemRole struct {
	ID     uint   `gorm:"primaryKey"`
	UserID string `gorm:"not null;index;size:64"`
	RoleID string `gorm:"not null;index;size:64"`
}

// GroupSystemRole assigns a role at system scope for a group.
type GroupSystemRole struct {
	ID      uint   `gorm:"primaryKey"`
	GroupID string `gorm:"not null;index;size:64"`
	RoleID  string `gorm:"not null;index;size:64"`
}
