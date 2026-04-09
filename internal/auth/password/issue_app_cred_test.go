package password

import (
	"testing"
	"time"

	"github.com/Kim-Hyo-Bin/gostone/internal/models"
	"github.com/Kim-Hyo-Bin/gostone/internal/testutil"
	"github.com/Kim-Hyo-Bin/gostone/internal/token"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func TestIssueAuthToken_applicationCredential_unrestricted(t *testing.T) {
	gdb, err := testutil.OpenMemory()
	if err != nil {
		t.Fatal(err)
	}
	fix, err := testutil.SeedAdmin(gdb, "userpw")
	if err != nil {
		t.Fatal(err)
	}
	mgr, err := token.NewManager(gdb, token.ProviderJWT, "jwt-appcred", time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte("app-secret"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	acID := uuid.NewString()
	ac := models.ApplicationCredential{
		ID:           acID,
		UserID:       fix.UserID,
		Name:         "cli",
		SecretHash:   string(hash),
		Unrestricted: true,
	}
	if err := gdb.Create(&ac).Error; err != nil {
		t.Fatal(err)
	}

	var req PasswordAuthRequest
	req.Auth.Identity.Methods = []string{"application_credential"}
	req.Auth.Identity.ApplicationCredential = &struct {
		ID     string `json:"id"`
		Secret string `json:"secret"`
	}{ID: acID, Secret: "app-secret"}

	tok, _, body, err := IssueAuthToken(gdb, mgr, &req)
	if err != nil {
		t.Fatal(err)
	}
	if tok == "" || body == nil {
		t.Fatal("empty response")
	}
	tokMap, _ := body["token"].(map[string]any)
	if tokMap == nil {
		t.Fatalf("body: %#v", body)
	}
	methods, _ := tokMap["methods"].([]string)
	if len(methods) != 1 || methods[0] != "application_credential" {
		t.Fatalf("methods %#v", tokMap["methods"])
	}
}
