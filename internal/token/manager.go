package token

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/fernet/fernet-go"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/Kim-Hyo-Bin/gostone/internal/models"
	"gorm.io/gorm"
)

// Manager selects between UUID (DB), JWT, or Keystone-compatible Fernet tokens.
type Manager struct {
	Provider        string
	DB              *gorm.DB
	JWT             *JWT
	FernetKeys      []*fernet.Key
	AuthMethodOrder []string
	TTL             time.Duration
}

const (
	ProviderUUID   = "uuid"
	ProviderJWT    = "jwt"
	ProviderFernet = "fernet"
)

// Issue issues a token with default auth method ["password"] when methods is nil.
func (m *Manager) Issue(userID, domainID, projectID string, roles []string) (string, time.Time, error) {
	return m.IssueWithMethods(userID, domainID, projectID, roles, nil)
}

// IssueWithMethods issues a token; if methods is empty, ["password"] is used.
func (m *Manager) IssueWithMethods(userID, domainID, projectID string, roles []string, methods []string) (string, time.Time, error) {
	if len(methods) == 0 {
		methods = []string{"password"}
	}
	switch strings.ToLower(strings.TrimSpace(m.Provider)) {
	case "", ProviderUUID:
		return m.issueUUID(userID, domainID, projectID, roles, methods)
	case ProviderJWT:
		if m.JWT == nil {
			return "", time.Time{}, errors.New("jwt token provider not configured")
		}
		return m.JWT.Issue(userID, domainID, projectID, roles, methods)
	case ProviderFernet:
		return m.issueFernet(userID, domainID, projectID, roles, methods)
	default:
		return "", time.Time{}, fmt.Errorf("unknown token provider %q", m.Provider)
	}
}

func (m *Manager) issueUUID(userID, domainID, projectID string, roles []string, methods []string) (string, time.Time, error) {
	if m.DB == nil {
		return "", time.Time{}, errors.New("database required for uuid token provider")
	}
	b, err := json.Marshal(struct {
		Roles    []string `json:"roles"`
		Methods  []string `json:"methods"`
		DomainID string   `json:"domain_id"`
		ProjID   string   `json:"project_id"`
	}{Roles: roles, Methods: methods, DomainID: domainID, ProjID: projectID})
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

func (m *Manager) issueFernet(userID, domainID, projectID string, roles []string, methods []string) (string, time.Time, error) {
	if len(m.FernetKeys) == 0 {
		return "", time.Time{}, errors.New("no fernet keys loaded")
	}
	primary := m.FernetKeys[0]
	now := time.Now()
	exp := now.Add(m.TTL)
	auditRand := uuid.New()
	auditStr := keystoneAuditEncode(auditRand[:])
	auditIDs := []string{auditStr}

	var plain []byte
	var err error
	if projectID == "" {
		plain, err = PackKeystoneFernetUnscoped(userID, methods, exp, auditIDs, m.AuthMethodOrder)
	} else {
		plain, err = PackKeystoneFernetProjectScoped(userID, projectID, methods, exp, auditIDs, m.AuthMethodOrder)
	}
	if err != nil {
		return "", time.Time{}, err
	}
	tok, err := fernetEncrypt(plain, primary)
	if err != nil {
		return "", time.Time{}, err
	}
	// Persist roles for middleware (Fernet payload does not carry role names).
	// Row primary key is a fixed-length hash (Fernet strings exceed varchar(64)).
	if m.DB != nil {
		rb, _ := json.Marshal(roles)
		row := models.AuthToken{
			ID:        fernetShadowID(tok),
			UserID:    userID,
			DomainID:  domainID,
			ProjectID: projectID,
			RolesJSON: string(rb),
			IssuedAt:  now,
			ExpiresAt: exp,
			RevokedAt: nil,
		}
		_ = m.DB.Create(&row).Error // best-effort shadow row for role lookup
	}
	return tok, exp, nil
}

func fernetShadowID(fernetToken string) string {
	sum := sha256.Sum256([]byte(fernetToken))
	return hex.EncodeToString(sum[:])
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
	case ProviderFernet:
		return m.parseFernet(tokenStr)
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
	roles := []string{}
	methods := []string{"password"}
	domainID := row.DomainID
	projectID := row.ProjectID
	if row.RolesJSON != "" {
		rj := strings.TrimSpace(row.RolesJSON)
		if strings.HasPrefix(rj, "{") {
			var meta struct {
				Roles    []string `json:"roles"`
				Methods  []string `json:"methods"`
				DomainID string   `json:"domain_id"`
				ProjID   string   `json:"project_id"`
			}
			if err := json.Unmarshal([]byte(row.RolesJSON), &meta); err == nil {
				roles = meta.Roles
				if len(meta.Methods) > 0 {
					methods = meta.Methods
				}
				if meta.DomainID != "" {
					domainID = meta.DomainID
				}
				if meta.ProjID != "" {
					projectID = meta.ProjID
				}
			}
		} else {
			_ = json.Unmarshal([]byte(row.RolesJSON), &roles)
		}
	}
	return &Claims{
		UserID:    row.UserID,
		DomainID:  domainID,
		ProjectID: projectID,
		Roles:     roles,
		Methods:   methods,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        row.ID,
			IssuedAt:  jwt.NewNumericDate(row.IssuedAt),
			ExpiresAt: jwt.NewNumericDate(row.ExpiresAt),
		},
	}, nil
}

func (m *Manager) parseFernet(tokenStr string) (*Claims, error) {
	plain, err := fernetDecrypt(tokenStr, m.FernetKeys)
	if err != nil {
		return nil, err
	}
	userID, projectID, methods, expPayload, err := UnpackKeystoneFernetPayload(plain, m.AuthMethodOrder)
	if err != nil {
		return nil, err
	}
	if !expPayload.IsZero() && time.Now().After(expPayload) {
		return nil, errors.New("token expired")
	}
	issued, err := FernetEnvelopeIssuedAt(tokenStr)
	if err != nil {
		issued = time.Now().UTC()
	}
	var roles []string
	domainID := ""
	if m.DB != nil {
		var row models.AuthToken
		if err := m.DB.Where("id = ?", fernetShadowID(tokenStr)).First(&row).Error; err == nil {
			if row.RevokedAt != nil {
				return nil, errors.New("token revoked")
			}
			_ = json.Unmarshal([]byte(row.RolesJSON), &roles)
		}
	}
	if domainID == "" && m.DB != nil {
		var u models.User
		if err := m.DB.Where("id = ?", userID).First(&u).Error; err == nil {
			domainID = u.DomainID
		}
	}
	return &Claims{
		UserID:    userID,
		DomainID:  domainID,
		ProjectID: projectID,
		Roles:     roles,
		Methods:   methods,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        tokenStr,
			IssuedAt:  jwt.NewNumericDate(issued),
			ExpiresAt: jwt.NewNumericDate(expPayload),
		},
	}, nil
}

// Revoke marks a UUID or Fernet-shadow token row revoked (no-op for JWT).
func (m *Manager) Revoke(tokenStr string) error {
	switch strings.ToLower(strings.TrimSpace(m.Provider)) {
	case ProviderJWT:
		return nil
	}
	if m.DB == nil {
		return nil
	}
	now := time.Now()
	id := tokenStr
	if strings.ToLower(strings.TrimSpace(m.Provider)) == ProviderFernet {
		id = fernetShadowID(tokenStr)
	}
	return m.DB.Model(&models.AuthToken{}).Where("id = ? AND revoked_at IS NULL", id).
		Update("revoked_at", now).Error
}
