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
	j := &token.JWT{Secret: []byte("k"), Issuer: "i", TTL: time.Hour}
	tokStr, _, err := j.Issue(fix.UserID, fix.DomainID, fix.ProjectID, []string{"admin"})
	if err != nil {
		t.Fatal(err)
	}
	claims, err := j.Parse(tokStr)
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

	_, err = BuildTokenResponse(gdb, &token.Claims{UserID: "missing-user-id"})
	if err == nil {
		t.Fatal("expected error")
	}
}
