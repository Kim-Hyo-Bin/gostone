package password

import (
	"strings"
	"testing"
	"time"

	"github.com/Kim-Hyo-Bin/gostone/internal/testutil"
	"github.com/Kim-Hyo-Bin/gostone/internal/token"
)

func TestIssuePasswordToken_errors(t *testing.T) {
	gdb, err := testutil.OpenMemory()
	if err != nil {
		t.Fatal(err)
	}
	j := &token.JWT{Secret: []byte("k"), Issuer: "i", TTL: time.Hour}

	_, _, _, err = IssuePasswordToken(gdb, j, nil)
	if err == nil {
		t.Fatal("nil request")
	}

	var req PasswordAuthRequest
	req.Auth.Identity.Methods = []string{"token"}
	_, _, _, err = IssuePasswordToken(gdb, j, &req)
	if err == nil || !strings.Contains(err.Error(), "unsupported") {
		t.Fatalf("got %v", err)
	}

	req.Auth.Identity.Methods = []string{"password"}
	_, _, _, err = IssuePasswordToken(gdb, j, &req)
	if err == nil || !strings.Contains(err.Error(), "password required") {
		t.Fatalf("got %v", err)
	}

	req.Auth.Identity.Password.User.Name = "admin"
	req.Auth.Identity.Password.User.Password = "secret"
	_, _, _, err = IssuePasswordToken(gdb, j, &req)
	if err == nil || !strings.Contains(err.Error(), "domain") {
		t.Fatalf("got %v", err)
	}
}

func TestIssuePasswordToken_success(t *testing.T) {
	gdb, err := testutil.OpenMemory()
	if err != nil {
		t.Fatal(err)
	}
	fix, err := testutil.SeedAdmin(gdb, "secret")
	if err != nil {
		t.Fatal(err)
	}
	j := &token.JWT{Secret: []byte("k"), Issuer: "i", TTL: time.Hour}

	req := PasswordAuthRequest{}
	req.Auth.Identity.Methods = []string{"password"}
	req.Auth.Identity.Password.User.Name = "admin"
	req.Auth.Identity.Password.User.Password = "secret"
	req.Auth.Identity.Password.User.Domain.Name = "Default"

	tok, _, body, err := IssuePasswordToken(gdb, j, &req)
	if err != nil {
		t.Fatal(err)
	}
	if tok == "" || body == nil {
		t.Fatal("empty token/body")
	}
	if body["token"] == nil {
		t.Fatal("missing token key")
	}
	// wrong password
	req.Auth.Identity.Password.User.Password = "nope"
	_, _, _, err = IssuePasswordToken(gdb, j, &req)
	if err == nil || !strings.Contains(err.Error(), "invalid password") {
		t.Fatalf("got %v", err)
	}

	// missing user in DB
	req.Auth.Identity.Password.User.Password = "secret"
	req.Auth.Identity.Password.User.Name = "nobody"
	_, _, _, err = IssuePasswordToken(gdb, j, &req)
	if err == nil || !strings.Contains(err.Error(), "user") {
		t.Fatalf("got %v", err)
	}

	// unknown domain name
	req.Auth.Identity.Password.User.Name = "admin"
	req.Auth.Identity.Password.User.Domain.Name = "Nope"
	_, _, _, err = IssuePasswordToken(gdb, j, &req)
	if err == nil || !strings.Contains(err.Error(), "domain") {
		t.Fatalf("got %v", err)
	}
	if fix.DomainID == "" {
		t.Fatal("fixture domain")
	}
}
