package conf

// Database is the [database] INI section.
type Database struct {
	Connection string `ini:"connection"`
	// MaxOpenConns caps open connections (0 = driver default).
	MaxOpenConns int `ini:"max_open_conns"`
	// MaxIdleConns caps idle pool size (0 = driver default).
	MaxIdleConns int `ini:"max_idle_conns"`
	// ConnMaxLifetimeSeconds closes connections after this age (0 = no limit).
	ConnMaxLifetimeSeconds int `ini:"conn_max_lifetime_seconds"`
}
