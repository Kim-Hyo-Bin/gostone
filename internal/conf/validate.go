package conf

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/Kim-Hyo-Bin/gostone/internal/token"
)

// Validate checks configuration that must hold before serving or migrating.
func Validate(c *Config) error {
	if c == nil {
		return fmt.Errorf("nil config")
	}
	if strings.TrimSpace(c.Database.Connection) == "" {
		return fmt.Errorf("[database] connection is required")
	}
	if c.Database.MaxOpenConns < 0 || c.Database.MaxIdleConns < 0 {
		return fmt.Errorf("[database] max_open_conns and max_idle_conns must be >= 0")
	}
	if c.Database.ConnMaxLifetimeSeconds < 0 {
		return fmt.Errorf("[database] conn_max_lifetime_seconds must be >= 0")
	}

	prov := strings.ToLower(strings.TrimSpace(c.Token.Provider))
	switch prov {
	case "", token.ProviderUUID:
		// ok
	case token.ProviderJWT:
		if strings.TrimSpace(c.Token.Secret) == "" {
			return fmt.Errorf("[token] secret is required when provider=jwt")
		}
	case token.ProviderFernet:
		if strings.TrimSpace(c.FernetTokens.KeyRepository) == "" {
			return fmt.Errorf("[fernet_tokens] key_repository is required when provider=fernet")
		}
	default:
		return fmt.Errorf("[token] unknown provider %q", c.Token.Provider)
	}

	if _, err := ListenBindings(&c.Service); err != nil {
		return err
	}

	if c.Service.EnforceAdminOnlyRoutes && strings.TrimSpace(c.Service.AdminOnlyPathPrefixes) == "" {
		return fmt.Errorf("[service] enforce_admin_only_routes requires admin_only_path_prefixes")
	}

	for _, u := range []struct {
		val   string
		label string
	}{
		{c.Service.PublicURL, "public_url"},
		{c.Service.AdminURL, "admin_url"},
		{c.Service.InternalURL, "internal_url"},
	} {
		if err := validateOptionalHTTPBaseURL(u.val, u.label); err != nil {
			return err
		}
	}
	return nil
}

func validateOptionalHTTPBaseURL(raw, label string) error {
	s := strings.TrimSpace(raw)
	if s == "" {
		return nil
	}
	u, err := url.Parse(s)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return fmt.Errorf("[service] %s must be a URL with scheme and host (e.g. http://controller:5000)", label)
	}
	scheme := strings.ToLower(u.Scheme)
	if scheme != "http" && scheme != "https" {
		return fmt.Errorf("[service] %s scheme must be http or https", label)
	}
	return nil
}
