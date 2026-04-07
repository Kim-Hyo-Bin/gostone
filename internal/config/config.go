// Package config loads gostone INI settings in the spirit of Keystone/oslo.config:
// --config-file / --config-dir, optional env overrides, and default paths under /etc/gostone/.
package config

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
}
