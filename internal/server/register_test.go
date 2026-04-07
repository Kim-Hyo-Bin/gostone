package server

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Kim-Hyo-Bin/gostone/internal/api/v3"
	"github.com/Kim-Hyo-Bin/gostone/internal/policy"
	"github.com/Kim-Hyo-Bin/gostone/internal/testutil"
	"github.com/Kim-Hyo-Bin/gostone/internal/token"
	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func newTestHub(t *testing.T) *v3.Hub {
	t.Helper()
	gdb, err := testutil.OpenMemory()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := testutil.SeedAdmin(gdb, "x"); err != nil {
		t.Fatal(err)
	}
	mgr, err := token.NewManager(gdb, token.ProviderJWT, "test-server-register", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	return &v3.Hub{DB: gdb, Tokens: mgr, Policy: policy.Default()}
}

func TestRegister_healthAndDiscovery(t *testing.T) {
	r := gin.New()
	Register(r, newTestHub(t))

	for _, tc := range []struct {
		path string
		code int
	}{
		{"/health", http.StatusOK},
		{"/", http.StatusMultipleChoices},
		{"/v3", http.StatusOK},
	} {
		t.Run(tc.path, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			req.Host = "127.0.0.1:5000"
			r.ServeHTTP(w, req)
			if w.Code != tc.code {
				t.Fatalf("%s: %d body %s", tc.path, w.Code, w.Body.String())
			}
		})
	}
}

func TestRegister_v3AuthTokensPOST(t *testing.T) {
	r := gin.New()
	Register(r, newTestHub(t))

	body := `{"auth":{"identity":{"methods":["password"],"password":{"user":{"name":"admin","password":"x","domain":{"name":"Default"}}}}}}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v3/auth/tokens", strings.NewReader(body))
	req.Host = "127.0.0.1:5000"
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("code %d %s", w.Code, w.Body.String())
	}
	if w.Header().Get("X-Subject-Token") == "" {
		t.Fatal("missing token header")
	}
}

func TestRegister_v3UsersWithToken(t *testing.T) {
	r := gin.New()
	Register(r, newTestHub(t))

	loginBody := `{"auth":{"identity":{"methods":["password"],"password":{"user":{"name":"admin","password":"x","domain":{"name":"Default"}}}}}}`
	w1 := httptest.NewRecorder()
	req1 := httptest.NewRequest(http.MethodPost, "/v3/auth/tokens", strings.NewReader(loginBody))
	req1.Host = "127.0.0.1:5000"
	req1.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w1, req1)
	if w1.Code != http.StatusCreated {
		t.Fatal(w1.Body.String())
	}
	token := w1.Header().Get("X-Subject-Token")

	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/v3/users", nil)
	req2.Host = "127.0.0.1:5000"
	req2.Header.Set("X-Auth-Token", token)
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Fatalf("code %d %s", w2.Code, w2.Body.String())
	}
}

func TestRegister_postUsers_create(t *testing.T) {
	gdb, err := testutil.OpenMemory()
	if err != nil {
		t.Fatal(err)
	}
	fix, err := testutil.SeedAdmin(gdb, "adminpw")
	if err != nil {
		t.Fatal(err)
	}
	mgr, err := token.NewManager(gdb, token.ProviderJWT, "post-users-test", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	hub := &v3.Hub{DB: gdb, Tokens: mgr, Policy: policy.Default()}
	r := gin.New()
	Register(r, hub)

	login := `{"auth":{"identity":{"methods":["password"],"password":{"user":{"name":"admin","password":"adminpw","domain":{"name":"Default"}}}}}}`
	w1 := httptest.NewRecorder()
	req1 := httptest.NewRequest(http.MethodPost, "/v3/auth/tokens", strings.NewReader(login))
	req1.Host = "127.0.0.1:5000"
	req1.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w1, req1)
	if w1.Code != http.StatusCreated {
		t.Fatal(w1.Body.String())
	}
	tok := w1.Header().Get("X-Subject-Token")

	body := fmt.Sprintf(`{"user":{"name":"alice","domain_id":%q,"password":"user-secret"}}`, fix.DomainID)
	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodPost, "/v3/users", strings.NewReader(body))
	req2.Host = "127.0.0.1:5000"
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("X-Auth-Token", tok)
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusCreated {
		t.Fatalf("code %d %s", w2.Code, w2.Body.String())
	}
}

func TestRegister_v3UsersForbiddenForMemberToken(t *testing.T) {
	h := newTestHub(t)
	r := gin.New()
	Register(r, h)
	// Token without admin role on list_users
	tok, _, err := h.Tokens.Issue("u1", "d1", "p1", []string{"member"})
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v3/users", nil)
	req.Host = "127.0.0.1:5000"
	req.Header.Set("X-Auth-Token", tok)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("code %d %s", w.Code, w.Body.String())
	}
}

func TestRegister_postAuthTokens_unsupportedMethod(t *testing.T) {
	r := gin.New()
	Register(r, newTestHub(t))
	body := `{"auth":{"identity":{"methods":["oauth1"],"password":{"user":{"name":"admin","password":"x","domain":{"name":"Default"}}}}}}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v3/auth/tokens", strings.NewReader(body))
	req.Host = "127.0.0.1:5000"
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("code %d %s", w.Code, w.Body.String())
	}
}

func TestRegister_postAuthTokens_malformedJSON(t *testing.T) {
	r := gin.New()
	Register(r, newTestHub(t))
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v3/auth/tokens", strings.NewReader(`{`))
	req.Host = "127.0.0.1:5000"
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("code %d %s", w.Code, w.Body.String())
	}
}

func TestRegister_getAuthTokens_missingHeader(t *testing.T) {
	r := gin.New()
	Register(r, newTestHub(t))
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v3/auth/tokens", nil)
	req.Host = "127.0.0.1:5000"
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("code %d", w.Code)
	}
}

func TestRegister_getAuthTokens_ok(t *testing.T) {
	h := newTestHub(t)
	r := gin.New()
	Register(r, h)
	loginBody := `{"auth":{"identity":{"methods":["password"],"password":{"user":{"name":"admin","password":"x","domain":{"name":"Default"}}}}}}`
	w1 := httptest.NewRecorder()
	req1 := httptest.NewRequest(http.MethodPost, "/v3/auth/tokens", strings.NewReader(loginBody))
	req1.Host = "127.0.0.1:5000"
	req1.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w1, req1)
	tok := w1.Header().Get("X-Subject-Token")

	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/v3/auth/tokens", nil)
	req2.Host = "127.0.0.1:5000"
	req2.Header.Set("X-Auth-Token", tok)
	r.ServeHTTP(w2, req2)
	if w2.Code != http.StatusOK {
		t.Fatalf("code %d %s", w2.Code, w2.Body.String())
	}
}

func TestRegister_headAuthTokens(t *testing.T) {
	h := newTestHub(t)
	r := gin.New()
	Register(r, h)
	tok, _, err := h.Tokens.Issue("u", "d", "p", []string{"admin"})
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodHead, "/v3/auth/tokens", nil)
	req.Host = "127.0.0.1:5000"
	req.Header.Set("X-Auth-Token", tok)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("code %d", w.Code)
	}
}

func TestRegister_deleteAuthTokens(t *testing.T) {
	h := newTestHub(t)
	r := gin.New()
	Register(r, h)
	tok, _, err := h.Tokens.Issue("u", "d", "p", []string{"admin"})
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/v3/auth/tokens", nil)
	req.Host = "127.0.0.1:5000"
	req.Header.Set("X-Auth-Token", tok)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNoContent {
		t.Fatalf("code %d", w.Code)
	}
}

func TestRegister_getUser_notFound(t *testing.T) {
	h := newTestHub(t)
	r := gin.New()
	Register(r, h)
	tok, _, err := h.Tokens.Issue("u", "d", "p", []string{"admin"})
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v3/users/00000000-0000-0000-0000-000000000001", nil)
	req.Host = "127.0.0.1:5000"
	req.Header.Set("X-Auth-Token", tok)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("code %d %s", w.Code, w.Body.String())
	}
}

func TestRegister_stubReturns501(t *testing.T) {
	h := newTestHub(t)
	r := gin.New()
	Register(r, h)
	tok, _, err := h.Tokens.Issue("u", "d", "p", []string{"admin"})
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v3/regions", nil)
	req.Host = "127.0.0.1:5000"
	req.Header.Set("X-Auth-Token", tok)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotImplemented {
		t.Fatalf("code %d %s", w.Code, w.Body.String())
	}
}
