package token

import "github.com/golang-jwt/jwt/v5"

// Claims is the in-process representation after validating any token provider.
// JSON tags match JWT claims issued by gostone (provider=jwt).
type Claims struct {
	UserID    string `json:"uid"`
	DomainID  string `json:"dom"`
	ProjectID string `json:"prj"`
	// ScopeDomainID is set for Keystone domain-scoped tokens (empty for project/unscoped).
	ScopeDomainID string   `json:"sdm,omitempty"`
	Roles         []string `json:"roles"`
	// Methods lists auth methods used to obtain the token (Keystone token body).
	Methods []string `json:"mth,omitempty"`

	// Optional Keystone Fernet/JWT extensions (populated when present in upstream tokens).
	TrustID            string   `json:"trt,omitempty"`
	SystemScope        string   `json:"sys,omitempty"`
	AppCredID          string   `json:"acd,omitempty"`
	AccessTokenID      string   `json:"oat,omitempty"`
	Thumbprint         string   `json:"tp,omitempty"`
	IdentityProviderID string   `json:"idp,omitempty"`
	ProtocolID         string   `json:"proto,omitempty"`
	FederatedGroupIDs  []string `json:"fgp,omitempty"`

	jwt.RegisteredClaims
}
