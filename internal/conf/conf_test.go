package conf

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	c := defaultConfig()
	if c.Service.Listen != ":5000" {
		t.Fatalf("listen: got %q", c.Service.Listen)
	}
	if c.Database.Connection == "" {
		t.Fatal("empty database connection")
	}
}

func TestResolveConfigPaths_explicitMissing(t *testing.T) {
	_, err := resolveConfigPaths(Options{ConfigFile: "/nonexistent/gostone-no-such.conf"})
	if err == nil || !strings.Contains(err.Error(), "not found") {
		t.Fatalf("expected not found error, got %v", err)
	}
}

func TestResolveConfigPaths_explicitOK(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "gostone.conf")
	if err := os.WriteFile(p, []byte("[service]\nlisten = :7001\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	files, err := resolveConfigPaths(Options{ConfigFile: p})
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 || files[0] != p {
		t.Fatalf("files: %#v", files)
	}
}

func TestFragmentConfFiles_missingDir(t *testing.T) {
	_, err := fragmentConfFiles(Options{ConfigDir: "/nonexistent/conf.d"})
	if err == nil || !strings.Contains(err.Error(), "not found") {
		t.Fatalf("got %v", err)
	}
}

func TestFragmentConfFiles_sorted(t *testing.T) {
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, "z.conf"), []byte("[service]\nlisten = :7002\n"), 0o600)
	_ = os.WriteFile(filepath.Join(dir, "a.conf"), []byte("[token]\nsecret = from-fragment\n"), 0o600)

	files, err := fragmentConfFiles(Options{ConfigDir: dir})
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 2 || filepath.Base(files[0]) != "a.conf" || filepath.Base(files[1]) != "z.conf" {
		t.Fatalf("want sorted a,z got %#v", files)
	}
}

func TestLoad_mergeFileAndFragment(t *testing.T) {
	mainDir := t.TempDir()
	main := filepath.Join(mainDir, "main.conf")
	if err := os.WriteFile(main, []byte(`
[service]
listen = :7003
[token]
secret = main-secret
expiration_hours = 1
`), 0o600); err != nil {
		t.Fatal(err)
	}
	fragDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(fragDir, "override.conf"), []byte(`
[service]
listen = :7004
`), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(Options{ConfigFile: main, ConfigDir: fragDir})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Service.Listen != ":7004" {
		t.Fatalf("fragment override: got %q", cfg.Service.Listen)
	}
	if cfg.Token.Secret != "main-secret" {
		t.Fatalf("token secret: got %q", cfg.Token.Secret)
	}
}

func TestLoad_noFilesUsesDefaults(t *testing.T) {
	t.Setenv("GOSTONE_CONFIG", "")
	// Ensure no accidental local gostone.conf is picked up in module root
	_ = os.Unsetenv("GOSTONE_CONFIG")
	cfg, err := Load(Options{})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Service.Listen != ":5000" {
		t.Fatalf("listen: %q", cfg.Service.Listen)
	}
}

func TestApplyEnvOverrides_viaLoad(t *testing.T) {
	t.Setenv("GOSTONE_HTTP_ADDR", ":9010")
	t.Setenv("GOSTONE_TOKEN_SECRET", "env-secret")
	t.Setenv("GOSTONE_SQLITE_DSN", "file::memory:?cache=shared")
	defer func() {
		_ = os.Unsetenv("GOSTONE_HTTP_ADDR")
		_ = os.Unsetenv("GOSTONE_TOKEN_SECRET")
		_ = os.Unsetenv("GOSTONE_SQLITE_DSN")
	}()

	cfg, err := Load(Options{})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Service.Listen != ":9010" || cfg.Token.Secret != "env-secret" {
		t.Fatalf("env: listen=%q secret=%q", cfg.Service.Listen, cfg.Token.Secret)
	}
}

func TestFileExistsAndDirExists(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "x")
	_ = os.WriteFile(f, []byte("a"), 0o600)
	if !fileExists(f) || fileExists(dir) {
		t.Fatal("fileExists")
	}
	if !dirExists(dir) || dirExists(f) {
		t.Fatal("dirExists")
	}
}
