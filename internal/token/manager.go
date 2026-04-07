package token

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Kim-Hyo-Bin/gostone/internal/models"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// NewManager builds a token manager. provider "" defaults to UUID (OpenStack-friendly).
func NewManager(db *gorm.DB, provider, secret string, ttl time.Duration) (*Manager, error) {
	p := strings.ToLower(strings.TrimSpace(provider))
	if p == "" {
		p = ProviderUUID
	}
	m := &Manager{Provider: p, DB: db, TTL: ttl}
	if p == ProviderJWT {
		if secret == "" {
			return nil, fmt.Errorf("token provider %q requires [token] secret", ProviderJWT)
		}
		m.JWT = &JWT{Secret: []byte(secret), Issuer: "gostone", TTL: ttl}
	}
	return m, nil
}

// Manager selects between JWT and persisted UUID tokens (Keystone UUID provider style).
type Manager struct {
	Provider string // "uuid" (default, OpenStack-friendly) or "jwt"
	DB       *gorm.DB
	JWT      *JWT
	TTL      time.Duration
}

const (
	ProviderUUID = "uuid"
	ProviderJWT  = "jwt"
)

// Issue creates a token string and expiry.
func (m *Manager) Issue(userID, domainID, projectID string, roles []string) (string, time.Time, error) {
	switch strings.ToLower(strings.TrimSpace(m.Provider)) {
	case "", ProviderUUID:
		return m.issueUUID(userID, domainID, projectID, roles)
	case ProviderJWT:
		if m.JWT == nil {
			return "", time.Time{}, errors.New("jwt token provider not configured")
		}
		return m.JWT.Issue(userID, domainID, projectID, roles)
	default:
		return "", time.Time{}, fmt.Errorf("unknown token provider %q", m.Provider)
	}
}

func (m *Manager) issueUUID(userID, domainID, projectID string, roles []string) (string, time.Time, error) {
	if m.DB == nil {
		return "", time.Time{}, errors.New("database required for uuid token provider")
	}
	b, err := json.Marshal(roles)
	if err != nil {
		return "", time.Time{}, err
	}
	now := time.Now()
	exp := now.Add(m.TTL)
	id := uuid.NewString()
	row := models.AuthToken{
		ID:        id,
		UserID:    userID,
		DomainID:  domainID,
		ProjectID: projectID,
		RolesJSON: string(b),
		IssuedAt:  now,
		ExpiresAt: exp,
		RevokedAt: nil,
	}
	if err := m.DB.Create(&row).Error; err != nil {
		return "", time.Time{}, err
	}
	return id, exp, nil
}

// Parse validates a token and returns claims (used by middleware and GET /v3/auth/tokens).
func (m *Manager) Parse(tokenStr string) (*Claims, error) {
	switch strings.ToLower(strings.TrimSpace(m.Provider)) {
	case "", ProviderUUID:
		return m.parseUUID(tokenStr)
	case ProviderJWT:
		if m.JWT == nil {
			return nil, errors.New("jwt token provider not configured")
		}
		return m.JWT.Parse(tokenStr)
	default:
		return nil, fmt.Errorf("unknown token provider %q", m.Provider)
	}
}

func (m *Manager) parseUUID(tokenStr string) (*Claims, error) {
	if m.DB == nil {
		return nil, errors.New("database required for uuid token provider")
	}
	var row models.AuthToken
	if err := m.DB.Where("id = ?", tokenStr).First(&row).Error; err != nil {
		return nil, err
	}
	if row.RevokedAt != nil {
		return nil, errors.New("token revoked")
	}
	if time.Now().After(row.ExpiresAt) {
		return nil, errors.New("token expired")
	}
	var roles []string
	if row.RolesJSON != "" {
		if err := json.Unmarshal([]byte(row.RolesJSON), &roles); err != nil {
			return nil, err
		}
	}
	return &Claims{
		UserID:    row.UserID,
		DomainID:  row.DomainID,
		ProjectID: row.ProjectID,
		Roles:     roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        row.ID,
			IssuedAt:  jwt.NewNumericDate(row.IssuedAt),
			ExpiresAt: jwt.NewNumericDate(row.ExpiresAt),
		},
	}, nil
}

// Revoke marks a UUID token revoked (no-op for JWT until blacklist exists).
func (m *Manager) Revoke(tokenStr string) error {
	if strings.ToLower(strings.TrimSpace(m.Provider)) == ProviderJWT {
		return nil
	}
	if m.DB == nil {
		return nil
	}
	now := time.Now()
	return m.DB.Model(&models.AuthToken{}).Where("id = ? AND revoked_at IS NULL", tokenStr).
		Update("revoked_at", now).Error
}
