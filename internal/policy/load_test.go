package policy

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Kim-Hyo-Bin/gostone/internal/auth"
)

func TestLoadFile_yamlAndRuleReference(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "rules.yaml")
	content := `
reader: "role:reader"
identity:list_users: "rule:reader or role:admin"
`
	if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	pol, err := LoadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	ctx := auth.Context{UserID: "u1", Roles: []string{"reader"}}
	if !pol.Allow("identity:list_users", ctx, nil) {
		t.Fatal("reader should allow list_users via rule:")
	}
	ctx2 := auth.Context{UserID: "u1", Roles: []string{}}
	if pol.Allow("identity:list_users", ctx2, nil) {
		t.Fatal("no role should deny")
	}
}

func TestLoadFile_jsonStillWorks(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "rules.json")
	if err := os.WriteFile(p, []byte(`{"identity:get_user":"role:admin"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	pol, err := LoadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	if pol.Rules["identity:get_user"] != "role:admin" {
		t.Fatalf("%q", pol.Rules["identity:get_user"])
	}
}

func TestEvalNotPrefix(t *testing.T) {
	p := &Policy{
		Rules: map[string]string{
			"x": "not role:admin",
		},
		DefaultRule: "false",
	}
	if p.Allow("x", auth.Context{UserID: "u", Roles: []string{"admin"}}, nil) {
		t.Fatal("admin should be denied")
	}
	if !p.Allow("x", auth.Context{UserID: "u", Roles: []string{"member"}}, nil) {
		t.Fatal("member should pass not role:admin")
	}
}

func TestEvalAndPrecedence(t *testing.T) {
	p := &Policy{
		Rules: map[string]string{
			"x": "role:admin and authenticated",
		},
		DefaultRule: "false",
	}
	if p.Allow("x", auth.Context{UserID: "", Roles: []string{"admin"}}, nil) {
		t.Fatal("not authenticated")
	}
	if !p.Allow("x", auth.Context{UserID: "u", Roles: []string{"admin"}}, nil) {
		t.Fatal("admin + authenticated")
	}
}
