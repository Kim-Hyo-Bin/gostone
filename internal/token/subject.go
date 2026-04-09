package token

import "errors"

// ErrConflictingScope means both ProjectID and ScopeDomainID were set.
var ErrConflictingScope = errors.New("token subject cannot set both project id and domain scope")

// TokenSubject is the authorization scope and roles stored inside issued tokens.
type TokenSubject struct {
	UserID   string
	DomainID string // user's home domain (identity domain)
	// ProjectID is non-empty for project-scoped tokens (Keystone project scope).
	ProjectID string
	// ScopeDomainID is non-empty for domain-scoped tokens (Keystone domain scope; Fernet v1).
	// It must be empty when ProjectID is set.
	ScopeDomainID string
	Roles         []string
	Methods       []string

	TrustID            string
	SystemScope        string
	AppCredID          string
	AccessTokenID      string
	Thumbprint         string
	IdentityProviderID string
	ProtocolID         string
	FederatedGroupIDs  []string
	// JTI, when non-empty, becomes JWT "jti" / UUID token id and the sole Keystone audit_ids entry.
	JTI string
}

func (s TokenSubject) normalizedMethods() []string {
	m := s.Methods
	if len(m) == 0 {
		m = []string{"password"}
	}
	return m
}

func (s TokenSubject) validate() error {
	if s.ProjectID != "" && s.ScopeDomainID != "" {
		return ErrConflictingScope
	}
	return nil
}
