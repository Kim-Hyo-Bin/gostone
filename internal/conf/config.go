// Package conf loads gostone INI settings (Keystone keystone.conf / oslo.config style).
// --config-file / --config-dir, optional env overrides, and default paths under /etc/gostone/.
package conf

// Config holds gostone settings mapped from INI sections (Keystone-like names where applicable).
type Config struct {
	Default struct {
		// Debug enables verbose Gin logging when true (similar in spirit to Keystone [DEFAULT] debug).
		Debug bool `ini:"debug"`
	} `ini:"DEFAULT"`

	Database struct {
		// Connection is the datastore URL/DSN (Keystone [database] connection: SQLAlchemy or native MySQL/Postgres/SQLite).
		Connection string `ini:"connection"`
	} `ini:"database"`

	Service struct {
		// Listen is the HTTP bind address (host:port or :port), e.g. ":5000".
		Listen string `ini:"listen"`
		// PublicURL is the base URL clients use (scheme://host:port), for catalog links and bootstrap identity endpoint. If empty, GOSTONE_PUBLIC_URL or http://127.0.0.1:<port> at bootstrap.
		PublicURL string `ini:"public_url"`
	} `ini:"service"`

	Token struct {
		// Provider is uuid (Keystone-style opaque DB tokens, default) or jwt (signed JWT for dev/tests).
		Provider string `ini:"provider"`
		// Secret is the HMAC key when provider=jwt (set [token] secret or GOSTONE_TOKEN_SECRET).
		Secret string `ini:"secret"`
		// ExpirationHours is token lifetime in hours.
		ExpirationHours int `ini:"expiration_hours"`
	} `ini:"token"`
}
