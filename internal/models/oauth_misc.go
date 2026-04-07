package models

import "time"

// OAuthConsumer is an OS-OAUTH1 consumer registration.
type OAuthConsumer struct {
	ID          string `gorm:"primaryKey;size:64"`
	Secret      string `gorm:"not null;type:text"`
	Description string `gorm:"type:text"`
}

// OAuth1AccessToken is a stored OAuth1 access token for a user.
type OAuth1AccessToken struct {
	ID         string `gorm:"primaryKey;size:64"`
	UserID     string `gorm:"not null;index;size:64"`
	ConsumerID string `gorm:"not null;index;size:64"`
	Secret     string `gorm:"not null;type:text"`
	Extra      string `gorm:"type:text"`
}

// ApplicationCredential is an application credential for a user.
type ApplicationCredential struct {
	ID           string     `gorm:"primaryKey;size:64"`
	UserID       string     `gorm:"not null;index;size:64"`
	Name         string     `gorm:"not null"`
	SecretHash   string     `gorm:"not null;type:text"`
	Description  string     `gorm:"type:text"`
	ExpiresAt    *time.Time `gorm:"index"`
	Unrestricted bool       `gorm:"default:false"`
}

// ApplicationCredentialRole binds a role to an application credential.
type ApplicationCredentialRole struct {
	ID        uint   `gorm:"primaryKey"`
	AppCredID string `gorm:"not null;index;size:64"`
	RoleID    string `gorm:"not null;index;size:64"`
}

// AccessRule limits paths for an application credential.
type AccessRule struct {
	ID        string `gorm:"primaryKey;size:64"`
	UserID    string `gorm:"not null;index;size:64"`
	Path      string `gorm:"not null;type:text"`
	Method    string `gorm:"not null"`
	ServiceID string `gorm:"type:text"`
}
