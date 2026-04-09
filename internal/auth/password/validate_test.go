package password

import (
	"testing"
	"time"

	"github.com/Kim-Hyo-Bin/gostone/internal/testutil"
	"github.com/Kim-Hyo-Bin/gostone/internal/token"
)

func TestBuildTokenResponse(t *testing.T) {
	gdb, err := testutil.OpenMemory()
	if err != nil {
		t.Fatal(err)
	}
	fix, err := testutil.SeedAdmin(gdb, "pw")
	if err != nil {
		t.Fatal(err)
	}
	mgr, err := token.NewManager(gdb, token.ProviderJWT, "k", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	tokStr, _, _, err := mgr.Issue(fix.UserID, fix.DomainID, fix.ProjectID, []string{"admin"})
	if err != nil {
		t.Fatal(err)
	}
	claims, err := mgr.Parse(tokStr)
	if err != nil {
		t.Fatal(err)
	}

	body, err := BuildTokenResponse(gdb, claims)
	if err != nil {
		t.Fatal(err)
	}
	if body["token"] == nil {
		t.Fatal(body)
	}
	tok, _ := body["token"].(map[string]any)
	roles, _ := tok["roles"].([]map[string]any)
	if len(roles) != 1 {
		t.Fatalf("roles: %#v", tok["roles"])
	}
	if roles[0]["id"] != fix.RoleID || roles[0]["name"] != "admin" {
		t.Fatalf("role object: %#v", roles[0])
	}

	_, err = BuildTokenResponse(gdb, &token.Claims{UserID: "missing-user-id"})
	if err == nil {
		t.Fatal("expected error")
	}
}
