package v3

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Kim-Hyo-Bin/gostone/internal/models"
	"github.com/Kim-Hyo-Bin/gostone/internal/policy"
	"github.com/Kim-Hyo-Bin/gostone/internal/testutil"
	"github.com/Kim-Hyo-Bin/gostone/internal/token"
	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func newTestHub(t *testing.T) *Hub {
	t.Helper()
	gdb, err := testutil.OpenMemory()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := testutil.SeedAdmin(gdb, "pass"); err != nil {
		t.Fatal(err)
	}
	j := &token.JWT{Secret: []byte("v3-smoke-test-secret-key"), Issuer: "i", TTL: time.Hour}
	return &Hub{DB: gdb, Tokens: j, Policy: policy.Default()}
}

func TestRegisterHealth(t *testing.T) {
	r := gin.New()
	RegisterHealth(r, newTestHub(t))
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req.Host = "127.0.0.1:5000"
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("code %d %s", w.Code, w.Body.String())
	}
}

func TestMount_passwordAuthAndListUsers(t *testing.T) {
	r := gin.New()
	h := newTestHub(t)
	Mount(r, h)

	body := `{"auth":{"identity":{"methods":["password"],"password":{"user":{"name":"admin","password":"pass","domain":{"name":"Default"}}}}}}`
	w1 := httptest.NewRecorder()
	req1 := httptest.NewRequest(http.MethodPost, "/v3/auth/tokens", strings.NewReader(body))
	req1.Host = "127.0.0.1:5000"
	req1.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w1, req1)
	if w1.Code != http.StatusCreated {
		t.Fatal(w1.Body.String())
	}
	tok := w1.Header().Get("X-Subject-Token")

	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/v3/users", nil)
	req2.Host = "127.0.0.1:5000"
	req2.Header.Set("X-Auth-Token", tok)
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Fatalf("code %d %s", w2.Code, w2.Body.String())
	}
}

func TestMount_listUsers_unauthenticated(t *testing.T) {
	r := gin.New()
	Mount(r, newTestHub(t))
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v3/users", nil)
	req.Host = "127.0.0.1:5000"
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("code %d", w.Code)
	}
}

func TestMount_getUser_selfWithoutAdminRole(t *testing.T) {
	r := gin.New()
	h := newTestHub(t)
	Mount(r, h)
	var u models.User
	if err := h.DB.Where("name = ?", "admin").First(&u).Error; err != nil {
		t.Fatal(err)
	}
	tok, _, err := h.Tokens.Issue(u.ID, u.DomainID, "", []string{})
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v3/users/"+u.ID, nil)
	req.Host = "127.0.0.1:5000"
	req.Header.Set("X-Auth-Token", tok)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("code %d %s", w.Code, w.Body.String())
	}
}
