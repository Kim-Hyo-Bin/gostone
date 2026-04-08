package token

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestFernetIssueParseRoundTrip(t *testing.T) {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i + 1)
	}
	enc := base64.URLEncoding.EncodeToString(key)
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "0"), []byte(enc), 0o600); err != nil {
		t.Fatal(err)
	}

	mgr, err := NewManagerWithConfig(Config{
		DB:            nil,
		Provider:      ProviderFernet,
		TTL:           time.Hour,
		FernetKeyRepo: dir,
		AuthMethods:   DefaultAuthMethods(),
	})
	if err != nil {
		t.Fatal(err)
	}

	tok, exp, err := mgr.IssueWithMethods("user-1", "dom-1", "proj-1", []string{"member"}, []string{"password"})
	if err != nil {
		t.Fatal(err)
	}
	if tok == "" {
		t.Fatal("empty token")
	}

	claims, err := mgr.Parse(tok)
	if err != nil {
		t.Fatal(err)
	}
	if claims.UserID != "user-1" {
		t.Fatalf("user: got %q", claims.UserID)
	}
	if claims.ProjectID != "proj-1" {
		t.Fatalf("project: got %q", claims.ProjectID)
	}
	if len(claims.Methods) != 1 || claims.Methods[0] != "password" {
		t.Fatalf("methods: %#v", claims.Methods)
	}
	if claims.ExpiresAt == nil || !claims.ExpiresAt.Time.Equal(exp.UTC().Truncate(time.Second)) {
		t.Fatalf("exp mismatch: claims=%v issue=%v", claims.ExpiresAt, exp)
	}
}
