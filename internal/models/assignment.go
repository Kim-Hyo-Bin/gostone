package models

// UserProjectRole assigns a role to a user on a project.
type UserProjectRole struct {
	ID        uint   `gorm:"primaryKey"`
	UserID    string `gorm:"not null;index:idx_upr_user"`
	RoleID    string `gorm:"not null;index:idx_upr_role"`
	ProjectID string `gorm:"not null;index:idx_upr_project"`
}
