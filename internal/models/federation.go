package models

// IdentityProvider is a SAML/OIDC-style IdP registration.
type IdentityProvider struct {
	ID          string `gorm:"primaryKey;size:64"`
	Enabled     bool   `gorm:"default:true"`
	Description string `gorm:"type:text"`
	RemoteIDs   string `gorm:"type:text"` // JSON array of strings
	Extra       string `gorm:"type:text"`
}

// FederationMapping maps remote attributes to local users/groups (JSON rules).
type FederationMapping struct {
	ID    string `gorm:"primaryKey;size:64"`
	Rules string `gorm:"not null;type:text"`
}

// ServiceProvider is a federation SP descriptor.
type ServiceProvider struct {
	ID          string `gorm:"primaryKey;size:64"`
	Description string `gorm:"type:text"`
	AuthURL     string `gorm:"type:text"`
	SpURL       string `gorm:"type:text"`
	Enabled     bool   `gorm:"default:true"`
}
