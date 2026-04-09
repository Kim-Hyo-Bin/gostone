package server

import (
	"encoding/json"
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
		{"/ready", http.StatusOK},
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

func TestRegister_users_patch_delete_head(t *testing.T) {
	gdb, err := testutil.OpenMemory()
	if err != nil {
		t.Fatal(err)
	}
	fix, err := testutil.SeedAdmin(gdb, "adminpw")
	if err != nil {
		t.Fatal(err)
	}
	mgr, err := token.NewManager(gdb, token.ProviderJWT, "patch-del-user-test", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	hub := &v3.Hub{DB: gdb, Tokens: mgr, Policy: policy.Default()}
	r := gin.New()
	Register(r, hub)

	loginAdmin := `{"auth":{"identity":{"methods":["password"],"password":{"user":{"name":"admin","password":"adminpw","domain":{"name":"Default"}}}}}}`
	w0 := httptest.NewRecorder()
	r0 := httptest.NewRequest(http.MethodPost, "/v3/auth/tokens", strings.NewReader(loginAdmin))
	r0.Host = "127.0.0.1:5000"
	r0.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w0, r0)
	if w0.Code != http.StatusCreated {
		t.Fatal(w0.Body.String())
	}
	adminTok := w0.Header().Get("X-Subject-Token")

	createBob := fmt.Sprintf(`{"user":{"name":"bob","domain_id":%q,"password":"bobpw"}}`, fix.DomainID)
	wc := httptest.NewRecorder()
	rc := httptest.NewRequest(http.MethodPost, "/v3/users", strings.NewReader(createBob))
	rc.Host = "127.0.0.1:5000"
	rc.Header.Set("Content-Type", "application/json")
	rc.Header.Set("X-Auth-Token", adminTok)
	r.ServeHTTP(wc, rc)
	if wc.Code != http.StatusCreated {
		t.Fatalf("create bob: %d %s", wc.Code, wc.Body.String())
	}
	var created struct {
		User struct {
			ID string `json:"id"`
		} `json:"user"`
	}
	if err := json.Unmarshal(wc.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	bobID := created.User.ID

	wh := httptest.NewRecorder()
	rh := httptest.NewRequest(http.MethodHead, "/v3/users/"+bobID, nil)
	rh.Host = "127.0.0.1:5000"
	rh.Header.Set("X-Auth-Token", adminTok)
	r.ServeHTTP(wh, rh)
	if wh.Code != http.StatusOK {
		t.Fatalf("HEAD %d", wh.Code)
	}

	wp := httptest.NewRecorder()
	rp := httptest.NewRequest(http.MethodPatch, "/v3/users/"+bobID, strings.NewReader(`{"user":{"name":"robert"}}`))
	rp.Host = "127.0.0.1:5000"
	rp.Header.Set("Content-Type", "application/json")
	rp.Header.Set("X-Auth-Token", adminTok)
	r.ServeHTTP(wp, rp)
	if wp.Code != http.StatusOK {
		t.Fatalf("PATCH %d %s", wp.Code, wp.Body.String())
	}

	bobLogin := `{"auth":{"identity":{"methods":["password"],"password":{"user":{"name":"robert","password":"bobpw","domain":{"name":"Default"}}}}}}`
	wl := httptest.NewRecorder()
	rl := httptest.NewRequest(http.MethodPost, "/v3/auth/tokens", strings.NewReader(bobLogin))
	rl.Host = "127.0.0.1:5000"
	rl.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(wl, rl)
	if wl.Code != http.StatusCreated {
		t.Fatalf("bob login: %d %s", wl.Code, wl.Body.String())
	}
	bobTok := wl.Header().Get("X-Subject-Token")

	wsp := httptest.NewRecorder()
	rsp := httptest.NewRequest(http.MethodPatch, "/v3/users/"+bobID, strings.NewReader(`{"user":{"password":"newbobpw"}}`))
	rsp.Host = "127.0.0.1:5000"
	rsp.Header.Set("Content-Type", "application/json")
	rsp.Header.Set("X-Auth-Token", bobTok)
	r.ServeHTTP(wsp, rsp)
	if wsp.Code != http.StatusOK {
		t.Fatalf("self PATCH password %d %s", wsp.Code, wsp.Body.String())
	}

	wd := httptest.NewRecorder()
	rd := httptest.NewRequest(http.MethodDelete, "/v3/users/"+bobID, nil)
	rd.Host = "127.0.0.1:5000"
	rd.Header.Set("X-Auth-Token", adminTok)
	r.ServeHTTP(wd, rd)
	if wd.Code != http.StatusNoContent {
		t.Fatalf("DELETE %d %s", wd.Code, wd.Body.String())
	}

	wg := httptest.NewRecorder()
	rg := httptest.NewRequest(http.MethodGet, "/v3/users/"+bobID, nil)
	rg.Host = "127.0.0.1:5000"
	rg.Header.Set("X-Auth-Token", adminTok)
	r.ServeHTTP(wg, rg)
	if wg.Code != http.StatusNotFound {
		t.Fatalf("get deleted: %d", wg.Code)
	}
}

func TestRegister_v3UsersForbiddenForMemberToken(t *testing.T) {
	h := newTestHub(t)
	r := gin.New()
	Register(r, h)
	// Token without admin role on list_users
	tok, _, _, err := h.Tokens.Issue("u1", "d1", "p1", []string{"member"})
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

func TestRegister_postAuthTokens_oauth1NotImplemented(t *testing.T) {
	r := gin.New()
	Register(r, newTestHub(t))
	body := `{"auth":{"identity":{"methods":["oauth1"],"password":{"user":{"name":"admin","password":"x","domain":{"name":"Default"}}}}}}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v3/auth/tokens", strings.NewReader(body))
	req.Host = "127.0.0.1:5000"
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotImplemented {
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
	tok, _, _, err := h.Tokens.Issue("u", "d", "p", []string{"admin"})
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
	tok, _, _, err := h.Tokens.Issue("u", "d", "p", []string{"admin"})
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
	tok, _, _, err := h.Tokens.Issue("u", "d", "p", []string{"admin"})
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
	tok, _, _, err := h.Tokens.Issue("u", "d", "p", []string{"admin"})
	if err != nil {
		t.Fatal(err)
	}
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v3/auth/tokens/OS-PKI/revoked", nil)
	req.Host = "127.0.0.1:5000"
	req.Header.Set("X-Auth-Token", tok)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotImplemented {
		t.Fatalf("code %d %s", w.Code, w.Body.String())
	}
}
