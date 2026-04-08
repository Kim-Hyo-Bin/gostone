package token

import (
	"strings"
	"testing"
	"time"
)

func TestJWT_IssueEmptySecret(t *testing.T) {
	j := &JWT{Secret: nil, Issuer: "t", TTL: time.Hour}
	_, _, err := j.Issue("u", "d", "p", []string{"admin"}, nil)
	if err == nil || !strings.Contains(err.Error(), "empty") {
		t.Fatalf("got %v", err)
	}
}

func TestJWT_RoundTrip(t *testing.T) {
	j := &JWT{Secret: []byte("test-secret-key-for-jwt"), Issuer: "gostone", TTL: time.Hour}
	tok, exp, err := j.Issue("user-1", "dom-1", "proj-1", []string{"admin", "member"}, nil)
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
	if claims.UserID != "user-1" || claims.DomainID != "dom-1" || claims.ProjectID != "proj-1" {
		t.Fatalf("%+v", claims)
	}
	if len(claims.Roles) != 2 {
		t.Fatal(claims.Roles)
	}
}

func TestJWT_ParseWrongSecret(t *testing.T) {
	j1 := &JWT{Secret: []byte("aaa"), Issuer: "i", TTL: time.Hour}
	tok, _, err := j1.Issue("u", "d", "", nil, nil)
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
