package token

import "github.com/golang-jwt/jwt/v5"

// Claims is the in-process representation after validating any token provider.
// JSON tags match JWT claims issued by gostone (provider=jwt).
type Claims struct {
	UserID    string   `json:"uid"`
	DomainID  string   `json:"dom"`
	ProjectID string   `json:"prj"`
	Roles     []string `json:"roles"`
	// Methods lists auth methods used to obtain the token (Keystone token body).
	Methods []string `json:"mth,omitempty"`
	jwt.RegisteredClaims
}
