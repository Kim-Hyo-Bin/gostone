package token

import (
	"strings"
	"testing"
	"time"
)

func TestJWT_IssueEmptySecret(t *testing.T) {
	j := &JWT{Secret: nil, Issuer: "t", TTL: time.Hour}
	_, _, err := j.Issue(TokenSubject{UserID: "u", DomainID: "d", ProjectID: "p", Roles: []string{"admin"}})
	if err == nil || !strings.Contains(err.Error(), "empty") {
		t.Fatalf("got %v", err)
	}
}

func TestJWT_RoundTrip(t *testing.T) {
	j := &JWT{Secret: []byte("test-secret-key-for-jwt"), Issuer: "gostone", TTL: time.Hour}
	tok, exp, err := j.Issue(TokenSubject{UserID: "user-1", DomainID: "dom-1", ProjectID: "proj-1", Roles: []string{"admin", "member"}})
	if err != nil {
		t.Fatal(err)
	}
	if !exp.After(time.Now()) {
		t.Fatal("expiry")
	}
	claims, err := j.Parse(tok)
	if err != nil {
		t.Fatal(err)
	}
	if claims.UserID != "user-1" || claims.DomainID != "dom-1" || claims.ProjectID != "proj-1" || claims.ScopeDomainID != "" {
		t.Fatalf("%+v", claims)
	}
	if len(claims.Roles) != 2 {
		t.Fatal(claims.Roles)
	}
}

func TestJWT_ParseWrongSecret(t *testing.T) {
	j1 := &JWT{Secret: []byte("aaa"), Issuer: "i", TTL: time.Hour}
	tok, _, err := j1.Issue(TokenSubject{UserID: "u", DomainID: "d", Roles: []string{}})
	if err != nil {
		t.Fatal(err)
	}
	j2 := &JWT{Secret: []byte("bbb"), Issuer: "i", TTL: time.Hour}
	_, err = j2.Parse(tok)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestJWT_ParseGarbage(t *testing.T) {
	j := &JWT{Secret: []byte("x"), Issuer: "i", TTL: time.Hour}
	_, err := j.Parse("not-a-valid-jwt")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestJWT_DomainScopeRoundTrip(t *testing.T) {
	j := &JWT{Secret: []byte("test-secret-key-domain-scope"), Issuer: "gostone", TTL: time.Hour}
	tok, _, err := j.Issue(TokenSubject{
		UserID:        "user-1",
		DomainID:      "dom-home",
		ScopeDomainID: "dom-scope",
		Roles:         []string{"reader"},
		Methods:       []string{"password"},
	})
	if err != nil {
		t.Fatal(err)
	}
	claims, err := j.Parse(tok)
	if err != nil {
		t.Fatal(err)
	}
	if claims.ScopeDomainID != "dom-scope" || claims.ProjectID != "" {
		t.Fatalf("%+v", claims)
	}
}
