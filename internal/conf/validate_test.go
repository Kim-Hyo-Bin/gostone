package conf

import (
	"strings"
	"testing"

	"github.com/Kim-Hyo-Bin/gostone/internal/token"
)

func TestValidate_okDefaults(t *testing.T) {
	c := defaultConfig()
	if err := Validate(c); err != nil {
		t.Fatal(err)
	}
}

func TestValidate_jwtNeedsSecret(t *testing.T) {
	c := defaultConfig()
	c.Token.Provider = token.ProviderJWT
	c.Token.Secret = ""
	if err := Validate(c); err == nil || !strings.Contains(err.Error(), "secret") {
		t.Fatalf("got %v", err)
	}
}

func TestValidate_fernetNeedsRepo(t *testing.T) {
	c := defaultConfig()
	c.Token.Provider = token.ProviderFernet
	c.FernetTokens.KeyRepository = ""
	if err := Validate(c); err == nil || !strings.Contains(err.Error(), "key_repository") {
		t.Fatalf("got %v", err)
	}
}

func TestValidate_enforceAdminOnlyNeedsPrefixes(t *testing.T) {
	c := defaultConfig()
	c.Service.EnforceAdminOnlyRoutes = true
	c.Service.AdminOnlyPathPrefixes = ""
	if err := Validate(c); err == nil {
		t.Fatal("expected error")
	}
}

func TestValidate_badPublicURL(t *testing.T) {
	c := defaultConfig()
	c.Service.PublicURL = "not-a-url"
	if err := Validate(c); err == nil {
		t.Fatal("expected error")
	}
}
