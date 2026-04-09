package db

import (
	"fmt"
	"net"
	"net/url"
	"strings"

	"github.com/glebarez/sqlite"
	"github.com/go-sql-driver/mysql"
	gmysql "gorm.io/driver/mysql"
	gpostgres "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Open returns a GORM DB from [database] connection (Keystone oslo.db–style URLs or native driver DSNs).
//
// Supported forms include:
//   - Keystone / SQLAlchemy: mysql+pymysql://..., mysql://..., postgresql+psycopg2://..., postgresql://..., postgres://..., sqlite:///...
//   - Go MySQL DSN: user:pass@tcp(host:3306)/dbname?parseTime=true&charset=utf8mb4
//   - PostgreSQL: postgres URL or libpq key=value (host=... user=... dbname=...)
//   - SQLite (glebarez): file:..., file::memory:..., or paths without scheme (treated as SQLite file path)
func Open(connection string) (*gorm.DB, error) {
	connection = strings.TrimSpace(connection)
	if connection == "" {
		return nil, fmt.Errorf("database connection is empty: set [database] connection in gostone.conf (Keystone-style) or GOSTONE_DATABASE_CONNECTION")
	}
	dialector, err := dialectorForConnection(connection)
	if err != nil {
		return nil, err
	}
	gdb, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("gorm open: %w", err)
	}
	return gdb, nil
}

func dialectorForConnection(connection string) (gorm.Dialector, error) {
	backend, dsn, err := normalizeConnection(connection)
	if err != nil {
		return nil, err
	}
	switch backend {
	case "sqlite":
		return sqlite.Open(dsn), nil
	case "mysql":
		return gmysql.Open(dsn), nil
	case "postgres":
		return gpostgres.Open(dsn), nil
	default:
		return nil, fmt.Errorf("unsupported database backend %q", backend)
	}
}

func normalizeConnection(s string) (backend string, dsn string, err error) {
	if s == "" {
		return "", "", fmt.Errorf("empty database connection string")
	}
	lower := strings.ToLower(s)

	switch {
	case strings.HasPrefix(lower, "mysql+pymysql://"), strings.HasPrefix(lower, "mysql://"),
		strings.HasPrefix(lower, "mariadb+pymysql://"), strings.HasPrefix(lower, "mariadb://"):
		dsn, err = sqlalchemyMySQLToDSN(s)
		return "mysql", dsn, err

	case strings.HasPrefix(lower, "postgresql+"), strings.HasPrefix(lower, "postgres+"):
		return "postgres", stripSQLAlchemyPostgresScheme(s), nil

	case strings.HasPrefix(lower, "postgresql://"), strings.HasPrefix(lower, "postgres://"):
		return "postgres", s, nil

	case strings.HasPrefix(lower, "sqlite:"):
		dsn, err = sqlalchemySQLiteToDSN(s)
		return "sqlite", dsn, err
	}

	// libpq-style (common in Postgres deployments)
	if strings.Contains(s, "host=") && strings.Contains(s, "dbname=") {
		return "postgres", s, nil
	}

	// Go MySQL DSN
	if strings.Contains(s, "@tcp(") {
		return "mysql", s, nil
	}

	// SQLite file / memory (no URL scheme)
	if strings.HasPrefix(lower, "file:") || strings.Contains(lower, "mode=memory") {
		return "sqlite", s, nil
	}

	// Plain path → SQLite file (relative or absolute without scheme)
	if !strings.Contains(s, "://") && !strings.Contains(s, "@") {
		return "sqlite", s, nil
	}

	return "", "", fmt.Errorf("unrecognized [database] connection; use mysql+pymysql://, postgresql://, sqlite:///, file:..., or native MySQL/Postgres DSN (see gostone.conf.example)")
}

func stripSQLAlchemyPostgresScheme(s string) string {
	idx := strings.Index(s, "://")
	if idx <= 0 {
		return s
	}
	scheme := strings.ToLower(s[:idx])
	rest := s[idx:]
	if plus := strings.Index(scheme, "+"); plus > 0 {
		base := scheme[:plus]
		if base == "postgresql" || base == "postgres" {
			return "postgres" + rest
		}
	}
	if scheme == "postgresql" {
		return "postgres" + rest
	}
	return s
}

func sqlalchemyMySQLToDSN(connection string) (string, error) {
	raw := connection
	switch {
	case strings.HasPrefix(strings.ToLower(raw), "mysql+pymysql://"):
		raw = "mysql://" + raw[len("mysql+pymysql://"):]
	case strings.HasPrefix(strings.ToLower(raw), "mariadb+pymysql://"):
		raw = "mysql://" + raw[len("mariadb+pymysql://"):]
	case strings.HasPrefix(strings.ToLower(raw), "mariadb://"):
		raw = "mysql://" + raw[len("mariadb://"):]
	}

	u, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("parse mysql connection URL: %w", err)
	}
	if u.Scheme != "mysql" {
		return "", fmt.Errorf("expected mysql URL scheme, got %q", u.Scheme)
	}

	pass, _ := u.User.Password()
	cfg := mysql.NewConfig()
	cfg.User = u.User.Username()
	cfg.Passwd = pass
	cfg.Net = "tcp"

	host := u.Hostname()
	port := u.Port()
	if host == "" {
		return "", fmt.Errorf("mysql connection URL must include a host (e.g. mysql+pymysql://user:pass@localhost/keystone)")
	}
	if port == "" {
		port = "3306"
	}
	cfg.Addr = net.JoinHostPort(host, port)

	dbName := strings.TrimPrefix(u.Path, "/")
	if dbName == "" {
		return "", fmt.Errorf("mysql connection URL must include a database name path (e.g. ...@host/keystone)")
	}
	cfg.DBName = dbName

	cfg.ParseTime = true
	cfg.Params = map[string]string{}
	for k, vals := range u.Query() {
		if len(vals) == 0 {
			continue
		}
		cfg.Params[k] = vals[0]
	}
	if _, ok := cfg.Params["charset"]; !ok {
		cfg.Params["charset"] = "utf8mb4"
	}

	return cfg.FormatDSN(), nil
}

func sqlalchemySQLiteToDSN(s string) (string, error) {
	// SQLAlchemy: sqlite:///relative.db, sqlite:////abs/path.db, sqlite:///:memory:
	const prefix = "sqlite:///"
	if len(s) < len(prefix) || !strings.EqualFold(s[:len(prefix)], prefix) {
		return "", fmt.Errorf("sqlite URL must start with sqlite:/// (SQLAlchemy style)")
	}
	body := s[len(prefix):]
	if body == ":memory:" {
		return "file::memory:?cache=shared", nil
	}
	if body == "" {
		return "", fmt.Errorf("sqlite URL missing path after sqlite:///")
	}
	// Absolute path on POSIX: sqlite://///tmp/x.db → body starts with / after fourth slash
	if strings.HasPrefix(body, "/") {
		return body, nil
	}
	return body, nil
}
