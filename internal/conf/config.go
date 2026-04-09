// Package conf loads gostone INI settings (Keystone keystone.conf / oslo.config style).
// --config-file / --config-dir, optional env overrides, and default paths under /etc/gostone/.
package conf

// Config holds gostone settings mapped from INI sections (Keystone-like names where applicable).
type Config struct {
	Default struct {
		// Debug enables verbose Gin logging when true (similar in spirit to Keystone [DEFAULT] debug).
		Debug bool `ini:"debug"`
	} `ini:"DEFAULT"`

	Database Database `ini:"database"`

	// Service is [service] (HTTP listeners, catalog URL, region).
	Service Service `ini:"service"`

	// Log controls access log formatting.
	Log Log `ini:"log"`

	Token struct {
		// Provider is uuid (Keystone-style opaque DB tokens, default) or jwt (signed JWT for dev/tests).
		Provider string `ini:"provider"`
		// Secret is the HMAC key when provider=jwt (set [token] secret or GOSTONE_TOKEN_SECRET).
		Secret string `ini:"secret"`
		// ExpirationHours is token lifetime in hours.
		ExpirationHours int `ini:"expiration_hours"`
	} `ini:"token"`

	// Auth mirrors Keystone [auth] (method order for Fernet bitmasks and token semantics).
	Auth struct {
		// Methods is a comma-separated list, e.g. "password,token" (default applied when empty).
		Methods string `ini:"methods"`
	} `ini:"auth"`

	// FernetTokens mirrors Keystone [fernet_tokens] when provider=fernet.
	FernetTokens struct {
		KeyRepository string `ini:"key_repository"`
	} `ini:"fernet_tokens"`

	// Policy holds optional JSON rule overrides (merged onto built-in defaults).
	Policy struct {
		File string `ini:"file"`
	} `ini:"policy"`

	// Discovery controls GET / and GET /v3 version advertisement (Identity API discovery).
	Discovery struct {
		APIVersionID string `ini:"api_version_id"`
		Updated      string `ini:"updated"`
		Status       string `ini:"status"`
	} `ini:"discovery"`
}
