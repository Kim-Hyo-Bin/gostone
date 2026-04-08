package conf

import "os"

// applyEnvOverrides applies legacy env vars when set (handy for dev / containers).
func applyEnvOverrides(c *Config) {
	if v := os.Getenv("GOSTONE_DATABASE_CONNECTION"); v != "" {
		c.Database.Connection = v
	} else if v := os.Getenv("GOSTONE_SQLITE_DSN"); v != "" {
		// Deprecated: use GOSTONE_DATABASE_CONNECTION (same as Keystone [database] connection).
		c.Database.Connection = v
	}
	if v := os.Getenv("GOSTONE_HTTP_ADDR"); v != "" {
		c.Service.Listen = v
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
	if v := os.Getenv("GOSTONE_AUTH_METHODS"); v != "" {
		c.Auth.Methods = v
	}
	if v := os.Getenv("GOSTONE_FERNET_KEY_REPO"); v != "" {
		c.FernetTokens.KeyRepository = v
	}
	if v := os.Getenv("GOSTONE_POLICY_FILE"); v != "" {
		c.Policy.File = v
	}
}
