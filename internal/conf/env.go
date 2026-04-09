package conf

import (
	"os"
	"strconv"
	"strings"
)

// applyEnvOverrides applies legacy env vars when set (handy for dev / containers).
func applyEnvOverrides(c *Config) {
	if v := os.Getenv("GOSTONE_DATABASE_CONNECTION"); v != "" {
		c.Database.Connection = v
	} else if v := os.Getenv("GOSTONE_SQLITE_DSN"); v != "" {
		// Deprecated: use GOSTONE_DATABASE_CONNECTION (same as Keystone [database] connection).
		c.Database.Connection = v
	}
	if v := os.Getenv("GOSTONE_DATABASE_MAX_OPEN_CONNS"); v != "" {
		if n, err := strconv.Atoi(strings.TrimSpace(v)); err == nil && n >= 0 {
			c.Database.MaxOpenConns = n
		}
	}
	if v := os.Getenv("GOSTONE_DATABASE_MAX_IDLE_CONNS"); v != "" {
		if n, err := strconv.Atoi(strings.TrimSpace(v)); err == nil && n >= 0 {
			c.Database.MaxIdleConns = n
		}
	}
	if v := os.Getenv("GOSTONE_DATABASE_CONN_MAX_LIFETIME_SECONDS"); v != "" {
		if n, err := strconv.Atoi(strings.TrimSpace(v)); err == nil && n >= 0 {
			c.Database.ConnMaxLifetimeSeconds = n
		}
	}
	if v := os.Getenv("GOSTONE_LOG_JSON"); v != "" {
		c.Log.JSON = parseEnvBool(v)
	}
	if v := os.Getenv("GOSTONE_HTTP_ADDR"); v != "" {
		c.Service.Listen = v
	}
	if v := os.Getenv("GOSTONE_LISTEN_PUBLIC"); v != "" {
		c.Service.ListenPublic = v
	}
	if v := os.Getenv("GOSTONE_LISTEN_ADMIN"); v != "" {
		c.Service.ListenAdmin = v
	}
	if v := os.Getenv("GOSTONE_LISTEN_INTERNAL"); v != "" {
		c.Service.ListenInternal = v
	}
	if v := os.Getenv("GOSTONE_TOKEN_SECRET"); v != "" {
		c.Token.Secret = v
	}
	if v := os.Getenv("GOSTONE_TOKEN_PROVIDER"); v != "" {
		c.Token.Provider = v
	}
	if v := os.Getenv("GOSTONE_PUBLIC_URL"); v != "" {
		c.Service.PublicURL = v
	}
	if v := os.Getenv("GOSTONE_ADMIN_URL"); v != "" {
		c.Service.AdminURL = v
	}
	if v := os.Getenv("GOSTONE_INTERNAL_URL"); v != "" {
		c.Service.InternalURL = v
	}
	if v := os.Getenv("GOSTONE_ENFORCE_ADMIN_ONLY_ROUTES"); v != "" {
		c.Service.EnforceAdminOnlyRoutes = parseEnvBool(v)
	}
	if v := os.Getenv("GOSTONE_ADMIN_ONLY_PATH_PREFIXES"); v != "" {
		c.Service.AdminOnlyPathPrefixes = v
	}
	if v := os.Getenv("GOSTONE_SHUTDOWN_TIMEOUT_SECONDS"); v != "" {
		if n, err := strconv.Atoi(strings.TrimSpace(v)); err == nil && n > 0 {
			c.Service.ShutdownTimeoutSeconds = n
		}
	}
	if v := os.Getenv("GOSTONE_TRUST_FORWARDED_HEADERS"); v != "" {
		c.Service.TrustForwardedHeaders = parseEnvBool(v)
	}
	if v := os.Getenv("GOSTONE_REGION_ID"); v != "" {
		c.Service.RegionID = v
	}
	if v := os.Getenv("GOSTONE_AUTH_METHODS"); v != "" {
		c.Auth.Methods = v
	}
	if v := os.Getenv("GOSTONE_FERNET_KEY_REPO"); v != "" {
		c.FernetTokens.KeyRepository = v
	}
	if v := os.Getenv("GOSTONE_POLICY_FILE"); v != "" {
		c.Policy.File = v
	}
	if v := os.Getenv("GOSTONE_DISCOVERY_API_VERSION_ID"); v != "" {
		c.Discovery.APIVersionID = v
	}
	if v := os.Getenv("GOSTONE_DISCOVERY_UPDATED"); v != "" {
		c.Discovery.Updated = v
	}
	if v := os.Getenv("GOSTONE_DISCOVERY_STATUS"); v != "" {
		c.Discovery.Status = v
	}
}

func parseEnvBool(v string) bool {
	v = strings.ToLower(strings.TrimSpace(v))
	return v == "1" || v == "true" || v == "yes" || v == "on"
}
