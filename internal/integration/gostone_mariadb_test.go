//go:build integration

package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Kim-Hyo-Bin/gostone/internal/api/v3"
	"github.com/Kim-Hyo-Bin/gostone/internal/auth/password"
	"github.com/Kim-Hyo-Bin/gostone/internal/bootstrap"
	"github.com/Kim-Hyo-Bin/gostone/internal/db"
	"github.com/Kim-Hyo-Bin/gostone/internal/policy"
	"github.com/Kim-Hyo-Bin/gostone/internal/server"
	"github.com/Kim-Hyo-Bin/gostone/internal/token"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func TestMariaDB_connectionAndMigrate(t *testing.T) {
	gdb := openIntegrationDB(t)
	if err := db.AutoMigrate(gdb); err != nil {
		t.Fatalf("AutoMigrate: %v", err)
	}
	sqlDB, err := gdb.DB()
	if err != nil {
		t.Fatal(err)
	}
	if err := sqlDB.Ping(); err != nil {
		t.Fatalf("ping: %v", err)
	}
}

func TestMariaDB_bootstrapAndPasswordAuth(t *testing.T) {
	gdb := openIntegrationDB(t)
	if err := db.AutoMigrate(gdb); err != nil {
		t.Fatalf("AutoMigrate: %v", err)
	}
	if err := bootstrap.EnsureIdentityCatalog(gdb); err != nil {
		t.Fatalf("EnsureIdentityCatalog: %v", err)
	}

	adminPW := os.Getenv("GOSTONE_BOOTSTRAP_ADMIN_PASSWORD")
	if adminPW == "" {
		adminPW = "admin"
	}
	t.Setenv("GOSTONE_BOOTSTRAP_ADMIN_PASSWORD", adminPW)

	if err := bootstrap.FromEnv(gdb); err != nil {
		t.Fatalf("FromEnv: %v", err)
	}

	secret := os.Getenv("GOSTONE_TOKEN_SECRET")
	if secret == "" {
		secret = "integration-test-jwt-secret"
	}
	prov := os.Getenv("GOSTONE_TOKEN_PROVIDER")
	if prov == "" {
		prov = token.ProviderJWT
	}
	mgr, err := token.NewManager(gdb, prov, secret, time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	req := password.PasswordAuthRequest{}
	req.Auth.Identity.Methods = []string{"password"}
	req.Auth.Identity.Password.User.Name = "admin"
	req.Auth.Identity.Password.User.Password = adminPW
	req.Auth.Identity.Password.User.Domain.Name = "Default"

	tok, _, body, err := password.IssuePasswordToken(gdb, mgr, &req)
	if err != nil {
		t.Fatalf("IssuePasswordToken: %v", err)
	}
	if tok == "" {
		t.Fatal("empty token")
	}
	if body == nil || body["token"] == nil {
		t.Fatalf("unexpected body: %#v", body)
	}
}

func TestMariaDB_httpListUsers(t *testing.T) {
	gin.SetMode(gin.TestMode)
	gdb := openIntegrationDB(t)
	if err := db.AutoMigrate(gdb); err != nil {
		t.Fatal(err)
	}
	if err := bootstrap.EnsureIdentityCatalog(gdb); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GOSTONE_BOOTSTRAP_ADMIN_PASSWORD", envOr("GOSTONE_BOOTSTRAP_ADMIN_PASSWORD", "admin"))
	if err := bootstrap.FromEnv(gdb); err != nil {
		t.Fatal(err)
	}

	secret := envOr("GOSTONE_TOKEN_SECRET", "integration-test-jwt-secret")
	mgr, err := token.NewManager(gdb, envOr("GOSTONE_TOKEN_PROVIDER", token.ProviderJWT), secret, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	hub := &v3.Hub{DB: gdb, Tokens: mgr, Policy: policy.Default()}
	r := gin.New()
	server.Register(r, hub)

	adminPW := envOr("GOSTONE_BOOTSTRAP_ADMIN_PASSWORD", "admin")
	login := `{"auth":{"identity":{"methods":["password"],"password":{"user":{"name":"admin","password":"` + jsonEscape(adminPW) + `","domain":{"name":"Default"}}}}}}`
	w0 := httptest.NewRecorder()
	r0 := httptest.NewRequest(http.MethodPost, "/v3/auth/tokens", strings.NewReader(login))
	r0.Host = "127.0.0.1:5000"
	r0.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w0, r0)
	if w0.Code != http.StatusCreated {
		t.Fatalf("auth status %d: %s", w0.Code, w0.Body.String())
	}
	subject := w0.Header().Get("X-Subject-Token")
	if subject == "" {
		t.Fatal("missing X-Subject-Token")
	}

	w1 := httptest.NewRecorder()
	r1 := httptest.NewRequest(http.MethodGet, "/v3/users", nil)
	r1.Host = "127.0.0.1:5000"
	r1.Header.Set("X-Auth-Token", subject)
	r.ServeHTTP(w1, r1)
	if w1.Code != http.StatusOK {
		t.Fatalf("list users %d: %s", w1.Code, w1.Body.String())
	}
	var wrap struct {
		Users []struct {
			Name string `json:"name"`
		} `json:"users"`
	}
	if err := json.Unmarshal(w1.Body.Bytes(), &wrap); err != nil {
		t.Fatal(err)
	}
	found := false
	for _, u := range wrap.Users {
		if u.Name == "admin" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected admin in users: %#v", wrap.Users)
	}
}

func openIntegrationDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := os.Getenv("GOSTONE_DATABASE_CONNECTION")
	if strings.TrimSpace(dsn) == "" {
		t.Fatal("GOSTONE_DATABASE_CONNECTION is required (e.g. mysql+pymysql://keystone:keystonepass@127.0.0.1:13306/gostone_integration)")
	}
	gdb, err := db.Open(dsn)
	if err != nil {
		t.Fatalf("db.Open: %v", err)
	}
	t.Cleanup(func() {
		sqlDB, _ := gdb.DB()
		if sqlDB != nil {
			_ = sqlDB.Close()
		}
	})
	return gdb
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func jsonEscape(s string) string {
	b, _ := json.Marshal(s)
	return string(b[1 : len(b)-1])
}
