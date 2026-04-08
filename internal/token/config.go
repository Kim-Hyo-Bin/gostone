package token

import (
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

// Config drives token provider construction (uuid, jwt, or fernet).
type Config struct {
	DB *gorm.DB

	Provider string
	TTL      time.Duration

	// JWTSecret is used when Provider == jwt.
	JWTSecret string

	// FernetKeyRepo is a directory of Keystone-style numbered key files when Provider == fernet.
	FernetKeyRepo string

	// AuthMethods is the ordered list matching Keystone [auth] methods (bitmask order).
	AuthMethods []string
}

// DefaultAuthMethods matches a minimal Keystone-style [auth] methods list (password, then token).
func DefaultAuthMethods() []string {
	return []string{"password", "token"}
}

// NewManager builds a Manager from DB + provider string + JWT secret (legacy signature).
func NewManager(db *gorm.DB, provider, jwtSecret string, ttl time.Duration) (*Manager, error) {
	return NewManagerWithConfig(Config{
		DB:            db,
		Provider:      provider,
		TTL:           ttl,
		JWTSecret:     jwtSecret,
		AuthMethods:   DefaultAuthMethods(),
		FernetKeyRepo: "",
	})
}

// NewManagerWithConfig is the full constructor (Fernet repo, auth method order).
func NewManagerWithConfig(c Config) (*Manager, error) {
	p := strings.ToLower(strings.TrimSpace(c.Provider))
	if p == "" {
		p = ProviderUUID
	}
	if c.TTL <= 0 {
		c.TTL = 24 * time.Hour
	}
	methods := c.AuthMethods
	if len(methods) == 0 {
		methods = DefaultAuthMethods()
	}
	m := &Manager{
		Provider:        p,
		DB:              c.DB,
		TTL:             c.TTL,
		AuthMethodOrder: methods,
	}
	switch p {
	case ProviderJWT:
		if c.JWTSecret == "" {
			return nil, fmt.Errorf("token provider %q requires [token] secret", ProviderJWT)
		}
		m.JWT = &JWT{Secret: []byte(c.JWTSecret), Issuer: "gostone", TTL: c.TTL}
	case ProviderFernet:
		if strings.TrimSpace(c.FernetKeyRepo) == "" {
			return nil, fmt.Errorf("token provider %q requires [fernet_tokens] key_repository", ProviderFernet)
		}
		keys, err := LoadFernetKeysFromRepo(strings.TrimSpace(c.FernetKeyRepo))
		if err != nil {
			return nil, err
		}
		m.FernetKeys = keys
	case ProviderUUID:
		if c.DB == nil {
			return nil, fmt.Errorf("token provider %q requires a database", ProviderUUID)
		}
	default:
		return nil, fmt.Errorf("unknown token provider %q", c.Provider)
	}
	return m, nil
}
