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
		// Connection is the datastore DSN (Keystone [database] connection; SQLite form for now).
		Connection string `ini:"connection"`
	} `ini:"database"`

	Service struct {
		// Listen is the HTTP bind address (host:port or :port), e.g. ":5000".
		Listen string `ini:"listen"`
	} `ini:"service"`

	Token struct {
		// Secret is the HMAC key for JWTs (set [token] secret or GOSTONE_TOKEN_SECRET).
		Secret string `ini:"secret"`
		// ExpirationHours is token lifetime in hours.
		ExpirationHours int `ini:"expiration_hours"`
	} `ini:"token"`
}
