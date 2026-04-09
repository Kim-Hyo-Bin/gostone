package conf

// Service holds [service] INI fields (Keystone-style public/admin/internal bind addresses).
type Service struct {
	// Listen is the default HTTP bind when not using role-specific listeners (host:port or :port).
	Listen string `ini:"listen"`
	// ListenPublic, ListenAdmin, ListenInternal enable multi-interface mode when any is non-empty.
	// Each non-empty value starts an additional http.Server with the same Gin engine and routes.
	ListenPublic   string `ini:"listen_public"`
	ListenAdmin    string `ini:"listen_admin"`
	ListenInternal string `ini:"listen_internal"`
	// PublicURL is the base URL clients use (scheme://host:port), for catalog links and bootstrap identity endpoint. If empty, GOSTONE_PUBLIC_URL or http://127.0.0.1:<port> at bootstrap.
	PublicURL string `ini:"public_url"`
	// AdminURL and InternalURL seed additional identity catalog endpoints (interface admin/internal) when non-empty.
	AdminURL    string `ini:"admin_url"`
	InternalURL string `ini:"internal_url"`
	// EnforceAdminOnlyRoutes with AdminOnlyPathPrefixes returns 403 on public/internal listeners for matching paths.
	EnforceAdminOnlyRoutes bool   `ini:"enforce_admin_only_routes"`
	AdminOnlyPathPrefixes  string `ini:"admin_only_path_prefixes"`
	ShutdownTimeoutSeconds int    `ini:"shutdown_timeout_seconds"`
	// TrustForwardedHeaders uses X-Forwarded-Host and X-Forwarded-Proto for discovery and WWW-Authenticate (trusted proxy only).
	TrustForwardedHeaders bool `ini:"trust_forwarded_headers"`
	// RegionID is the default region id when seeding an empty service catalog (matches bootstrap --bootstrap-region-id). Empty means RegionOne.
	RegionID string `ini:"region_id"`
}
