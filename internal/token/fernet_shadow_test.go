package token

import (
	"encoding/base64"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Kim-Hyo-Bin/gostone/internal/models"
	"github.com/Kim-Hyo-Bin/gostone/internal/testutil"
	"github.com/google/uuid"
)

func TestFernetParseRequiresShadowWhenDB(t *testing.T) {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i + 3)
	}
	enc := base64.URLEncoding.EncodeToString(key)
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "0"), []byte(enc), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg := Config{
		Provider:      ProviderFernet,
		TTL:           time.Hour,
		FernetKeyRepo: dir,
		AuthMethods:   DefaultAuthMethods(),
	}

	mgrNoDB, err := NewManagerWithConfig(cfg)
	if err != nil {
		t.Fatal(err)
	}
	tok, _, err := mgrNoDB.IssueWithMethods("u1", "d1", "p1", []string{"member"}, []string{"password"})
	if err != nil {
		t.Fatal(err)
	}

	gdb, err := testutil.OpenMemory()
	if err != nil {
		t.Fatal(err)
	}
	cfg.DB = gdb
	mgrDB, err := NewManagerWithConfig(cfg)
	if err != nil {
		t.Fatal(err)
	}
	_, err = mgrDB.Parse(tok)
	if !errors.Is(err, ErrFernetShadowMissing) {
		t.Fatalf("want ErrFernetShadowMissing, got %v", err)
	}
}

func TestFernetIssueParseWithDBUsesShadow(t *testing.T) {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i + 7)
	}
	enc := base64.URLEncoding.EncodeToString(key)
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "0"), []byte(enc), 0o600); err != nil {
		t.Fatal(err)
	}
	gdb, err := testutil.OpenMemory()
	if err != nil {
		t.Fatal(err)
	}
	domainID := uuid.NewString()
	userID := uuid.NewString()
	if err := gdb.Create(&models.Domain{ID: domainID, Name: "D", Enabled: true}).Error; err != nil {
		t.Fatal(err)
	}
	if err := gdb.Create(&models.User{
		ID: userID, DomainID: domainID, Name: "u", Enabled: true, PasswordHash: "x",
	}).Error; err != nil {
		t.Fatal(err)
	}

	mgr, err := NewManagerWithConfig(Config{
		DB:            gdb,
		Provider:      ProviderFernet,
		TTL:           time.Hour,
		FernetKeyRepo: dir,
		AuthMethods:   DefaultAuthMethods(),
	})
	if err != nil {
		t.Fatal(err)
	}
	tok, _, err := mgr.IssueWithMethods(userID, domainID, "proj-x", []string{"admin"}, []string{"password"})
	if err != nil {
		t.Fatal(err)
	}
	claims, err := mgr.Parse(tok)
	if err != nil {
		t.Fatal(err)
	}
	if claims.UserID != userID || claims.DomainID != domainID {
		t.Fatalf("claims %+v", claims)
	}
	if len(claims.Roles) != 1 || claims.Roles[0] != "admin" {
		t.Fatalf("roles %+v", claims.Roles)
	}
	_ = tok
}
