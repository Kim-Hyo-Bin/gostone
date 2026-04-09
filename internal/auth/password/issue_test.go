package password

import (
	"strings"
	"testing"
	"time"

	"github.com/Kim-Hyo-Bin/gostone/internal/models"
	"github.com/Kim-Hyo-Bin/gostone/internal/testutil"
	"github.com/Kim-Hyo-Bin/gostone/internal/token"
	"github.com/google/uuid"
)

func TestIssuePasswordToken_errors(t *testing.T) {
	gdb, err := testutil.OpenMemory()
	if err != nil {
		t.Fatal(err)
	}
	mgr, err := token.NewManager(gdb, token.ProviderJWT, "k", time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	_, _, _, err = IssuePasswordToken(gdb, mgr, nil)
	if err == nil {
		t.Fatal("nil request")
	}

	var req PasswordAuthRequest
	req.Auth.Identity.Methods = []string{"token"}
	_, _, _, err = IssuePasswordToken(gdb, mgr, &req)
	if err == nil || !strings.Contains(err.Error(), "unsupported") {
		t.Fatalf("got %v", err)
	}

	req.Auth.Identity.Methods = []string{"password"}
	_, _, _, err = IssuePasswordToken(gdb, mgr, &req)
	if err == nil || !strings.Contains(err.Error(), "password required") {
		t.Fatalf("got %v", err)
	}

	req.Auth.Identity.Password.User.Name = "admin"
	req.Auth.Identity.Password.User.Password = "secret"
	_, _, _, err = IssuePasswordToken(gdb, mgr, &req)
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
	mgr, err := token.NewManager(gdb, token.ProviderJWT, "k", time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	req := PasswordAuthRequest{}
	req.Auth.Identity.Methods = []string{"password"}
	req.Auth.Identity.Password.User.Name = "admin"
	req.Auth.Identity.Password.User.Password = "secret"
	req.Auth.Identity.Password.User.Domain.Name = "Default"

	tok, _, body, err := IssuePasswordToken(gdb, mgr, &req)
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
	_, _, _, err = IssuePasswordToken(gdb, mgr, &req)
	if err == nil || !strings.Contains(err.Error(), "invalid password") {
		t.Fatalf("got %v", err)
	}

	// missing user in DB
	req.Auth.Identity.Password.User.Password = "secret"
	req.Auth.Identity.Password.User.Name = "nobody"
	_, _, _, err = IssuePasswordToken(gdb, mgr, &req)
	if err == nil || !strings.Contains(err.Error(), "user") {
		t.Fatalf("got %v", err)
	}

	// unknown domain name
	req.Auth.Identity.Password.User.Name = "admin"
	req.Auth.Identity.Password.User.Domain.Name = "Nope"
	_, _, _, err = IssuePasswordToken(gdb, mgr, &req)
	if err == nil || !strings.Contains(err.Error(), "domain") {
		t.Fatalf("got %v", err)
	}
	if fix.DomainID == "" {
		t.Fatal("fixture domain")
	}
}

func TestIssuePasswordToken_domainScope(t *testing.T) {
	gdb, err := testutil.OpenMemory()
	if err != nil {
		t.Fatal(err)
	}
	fix, err := testutil.SeedAdmin(gdb, "secret")
	if err != nil {
		t.Fatal(err)
	}
	readerID := uuid.NewString()
	if err := gdb.Create(&models.Role{ID: readerID, Name: "reader", DomainID: ""}).Error; err != nil {
		t.Fatal(err)
	}
	if err := gdb.Create(&models.UserDomainRole{UserID: fix.UserID, DomainID: fix.DomainID, RoleID: readerID}).Error; err != nil {
		t.Fatal(err)
	}
	mgr, err := token.NewManager(gdb, token.ProviderJWT, "k", time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	req := PasswordAuthRequest{}
	req.Auth.Identity.Methods = []string{"password"}
	req.Auth.Identity.Password.User.Name = "admin"
	req.Auth.Identity.Password.User.Password = "secret"
	req.Auth.Identity.Password.User.Domain.Name = "Default"
	scopeDom := struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}{ID: fix.DomainID}
	req.Auth.Scope = &AuthScope{Domain: &scopeDom}

	tok, _, body, err := IssuePasswordToken(gdb, mgr, &req)
	if err != nil {
		t.Fatal(err)
	}
	tokMap, _ := body["token"].(map[string]any)
	if tokMap == nil {
		t.Fatal("token map")
	}
	if tokMap["project_scope"] != false {
		t.Fatalf("project_scope: %v", tokMap["project_scope"])
	}
	if tokMap["domain"] == nil {
		t.Fatal("expected token.domain for domain scope")
	}
	claims, err := mgr.Parse(tok)
	if err != nil {
		t.Fatal(err)
	}
	if claims.ScopeDomainID != fix.DomainID || claims.ProjectID != "" {
		t.Fatalf("claims %+v", claims)
	}
	if tok == "" {
		t.Fatal("token")
	}
}

func TestIssueAuthToken_ambiguousScope(t *testing.T) {
	gdb, err := testutil.OpenMemory()
	if err != nil {
		t.Fatal(err)
	}
	_, err = testutil.SeedAdmin(gdb, "secret")
	if err != nil {
		t.Fatal(err)
	}
	mgr, err := token.NewManager(gdb, token.ProviderJWT, "k", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	var req PasswordAuthRequest
	req.Auth.Identity.Methods = []string{"password"}
	req.Auth.Identity.Password.User.Name = "admin"
	req.Auth.Identity.Password.User.Password = "secret"
	req.Auth.Identity.Password.User.Domain.Name = "Default"
	d := struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}{Name: "Default"}
	p := struct {
		ID     string `json:"id"`
		Name   string `json:"name"`
		Domain *struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"domain"`
	}{Name: "admin", Domain: &struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}{Name: "Default"}}
	req.Auth.Scope = &AuthScope{Domain: &d, Project: &p}
	_, _, _, err = IssueAuthToken(gdb, mgr, &req)
	if err == nil || !strings.Contains(err.Error(), "ambiguous scope") {
		t.Fatalf("got %v", err)
	}
}

func TestIssueAuthToken_passwordTotpCompositeNotImplemented(t *testing.T) {
	gdb, err := testutil.OpenMemory()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := testutil.SeedAdmin(gdb, "secret"); err != nil {
		t.Fatal(err)
	}
	mgr, err := token.NewManager(gdb, token.ProviderJWT, "k", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	var req PasswordAuthRequest
	req.Auth.Identity.Methods = []string{"password", "totp"}
	req.Auth.Identity.Password.User.Name = "admin"
	req.Auth.Identity.Password.User.Password = "secret"
	req.Auth.Identity.Password.User.Domain.Name = "Default"
	_, _, _, err = IssueAuthToken(gdb, mgr, &req)
	if err == nil || !strings.Contains(err.Error(), "not implemented") {
		t.Fatalf("got %v", err)
	}
}

func TestIssueAuthToken_passwordByUserIDWithoutDomain(t *testing.T) {
	gdb, err := testutil.OpenMemory()
	if err != nil {
		t.Fatal(err)
	}
	fix, err := testutil.SeedAdmin(gdb, "secret")
	if err != nil {
		t.Fatal(err)
	}
	mgr, err := token.NewManager(gdb, token.ProviderJWT, "k", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	var req PasswordAuthRequest
	req.Auth.Identity.Methods = []string{"password"}
	req.Auth.Identity.Password.User.ID = fix.UserID
	req.Auth.Identity.Password.User.Password = "secret"
	tok, _, body, err := IssueAuthToken(gdb, mgr, &req)
	if err != nil {
		t.Fatal(err)
	}
	if tok == "" || body == nil || body["token"] == nil {
		t.Fatalf("unexpected response: %#v", body)
	}
}

func TestIssueAuthToken_singleTotpNotImplemented(t *testing.T) {
	gdb, err := testutil.OpenMemory()
	if err != nil {
		t.Fatal(err)
	}
	mgr, err := token.NewManager(gdb, token.ProviderJWT, "k", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	var req PasswordAuthRequest
	req.Auth.Identity.Methods = []string{"totp"}
	_, _, _, err = IssueAuthToken(gdb, mgr, &req)
	if err == nil || !strings.Contains(err.Error(), "not implemented") {
		t.Fatalf("got %v", err)
	}
}
