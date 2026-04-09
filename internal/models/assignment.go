package models

// UserProjectRole assigns a role to a user on a project.
type UserProjectRole struct {
	ID        uint   `gorm:"primaryKey"`
	UserID    string `gorm:"not null;index:idx_upr_user;size:64"`
	RoleID    string `gorm:"not null;index:idx_upr_role;size:64"`
	ProjectID string `gorm:"not null;index:idx_upr_project;size:64"`
}
