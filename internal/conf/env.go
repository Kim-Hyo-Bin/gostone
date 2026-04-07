package conf

import "os"

// applyEnvOverrides applies legacy env vars when set (handy for dev / containers).
func applyEnvOverrides(c *Config) {
	if v := os.Getenv("GOSTONE_SQLITE_DSN"); v != "" {
		c.Database.Connection = v
	}
	if v := os.Getenv("GOSTONE_HTTP_ADDR"); v != "" {
		c.Service.Listen = v
	}
	if v := os.Getenv("GOSTONE_TOKEN_SECRET"); v != "" {
		c.Token.Secret = v
	}
}
