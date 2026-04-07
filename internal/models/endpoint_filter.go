package models

// EndpointPolicyLink associates an endpoint with an identity policy document.
type EndpointPolicyLink struct {
	ID         uint   `gorm:"primaryKey"`
	EndpointID string `gorm:"not null;index:idx_ep_pol,priority:1;uniqueIndex:idx_ep_pol_pair;size:64"`
	PolicyID   string `gorm:"not null;index;uniqueIndex:idx_ep_pol_pair;size:64"`
}

// ProjectEndpointFilter links a project to an allowed endpoint (OS-EP-FILTER).
type ProjectEndpointFilter struct {
	ID         uint   `gorm:"primaryKey"`
	ProjectID  string `gorm:"not null;index:idx_pef_proj;size:64"`
	EndpointID string `gorm:"not null;index;size:64"`
}

// EndpointGroup is a named group of endpoints for filtering.
type EndpointGroup struct {
	ID          string `gorm:"primaryKey;size:64"`
	Name        string `gorm:"not null;size:255"`
	Description string `gorm:"type:text"`
}

// EndpointGroupMember links an endpoint to an endpoint group.
type EndpointGroupMember struct {
	ID              uint   `gorm:"primaryKey"`
	EndpointGroupID string `gorm:"not null;index;size:64"`
	EndpointID      string `gorm:"not null;index;size:64"`
}

// EndpointGroupProject links an endpoint group to a project.
type EndpointGroupProject struct {
	ID              uint   `gorm:"primaryKey"`
	EndpointGroupID string `gorm:"not null;index;size:64"`
	ProjectID       string `gorm:"not null;index;size:64"`
}
