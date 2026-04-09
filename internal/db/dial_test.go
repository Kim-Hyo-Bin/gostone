package db

import (
	"strings"
	"testing"
)

func TestNormalizeConnection_sqlite(t *testing.T) {
	for _, tc := range []struct {
		in       string
		wantBack string
		wantPre  string // prefix of dsn
	}{
		{"file::memory:?cache=shared", "sqlite", "file::memory:"},
		{"./local.db", "sqlite", "./local.db"},
		{"sqlite:///:memory:", "sqlite", "file::memory:"},
		{"sqlite:///rel.sqlite", "sqlite", "rel.sqlite"},
		{"sqlite:////tmp/abs.sqlite", "sqlite", "/tmp/abs.sqlite"},
	} {
		b, d, err := normalizeConnection(tc.in)
		if err != nil {
			t.Fatalf("%q: %v", tc.in, err)
		}
		if b != tc.wantBack || !strings.HasPrefix(d, tc.wantPre) {
			t.Fatalf("%q -> %s %q want prefix %q", tc.in, b, d, tc.wantPre)
		}
	}
}

func TestNormalizeConnection_mysqlSQLAlchemy(t *testing.T) {
	b, d, err := normalizeConnection("mysql+pymysql://u:p@db.example:3307/keystone?charset=utf8mb4")
	if err != nil {
		t.Fatal(err)
	}
	if b != "mysql" {
		t.Fatalf("backend %s", b)
	}
	if !strings.Contains(d, "u:p@tcp(db.example:3307)/keystone") {
		t.Fatalf("dsn %q", d)
	}
	if !strings.Contains(d, "parseTime=true") {
		t.Fatalf("parseTime: %q", d)
	}
}

func TestNormalizeConnection_mysqlNative(t *testing.T) {
	b, d, err := normalizeConnection("root:secret@tcp(127.0.0.1:3306)/keystone?parseTime=true")
	if err != nil || b != "mysql" || d == "" {
		t.Fatalf("%s %s %v", b, d, err)
	}
}

func TestNormalizeConnection_postgres(t *testing.T) {
	for _, in := range []string{
		"postgres://u:p@localhost:5432/keystone?sslmode=disable",
		"postgresql://u:p@localhost/keystone",
		"postgresql+psycopg2://u:p@h/db",
		"host=localhost port=5432 user=u password=p dbname=keystone sslmode=disable",
	} {
		b, d, err := normalizeConnection(in)
		if err != nil {
			t.Fatalf("%q: %v", in, err)
		}
		if b != "postgres" {
			t.Fatalf("%q -> %s", in, b)
		}
		if d == "" {
			t.Fatal(in)
		}
	}
	if got := stripSQLAlchemyPostgresScheme("postgresql+psycopg2://x:y@h/db"); !strings.HasPrefix(got, "postgres://") {
		t.Fatalf("%q", got)
	}
}

func TestNormalizeConnection_errors(t *testing.T) {
	for _, in := range []string{
		"",
		"http://example.com/db",
		"mysql+pymysql://u:p@/nodbpath",
	} {
		_, _, err := normalizeConnection(in)
		if err == nil {
			t.Fatalf("expected error for %q", in)
		}
	}
}

func TestSqlalchemyMySQL_missingHost(t *testing.T) {
	_, err := sqlalchemyMySQLToDSN("mysql://user:pass@/dbname")
	if err == nil {
		t.Fatal("expected error")
	}
}
