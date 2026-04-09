package v3

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Kim-Hyo-Bin/gostone/internal/policy"
	"github.com/Kim-Hyo-Bin/gostone/internal/testutil"
	"github.com/Kim-Hyo-Bin/gostone/internal/token"
	"github.com/gin-gonic/gin"
)

func TestCatalog_regionsServicesEndpoints(t *testing.T) {
	gdb, err := testutil.OpenMemory()
	if err != nil {
		t.Fatal(err)
	}
	fix, err := testutil.SeedAdmin(gdb, "pw")
	if err != nil {
		t.Fatal(err)
	}
	mgr, err := token.NewManager(gdb, token.ProviderJWT, "catalog-test", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	h := &Hub{DB: gdb, Tokens: mgr, Policy: policy.Default()}
	r := gin.New()
	Mount(r, h)

	login := `{"auth":{"identity":{"methods":["password"],"password":{"user":{"name":"admin","password":"pw","domain":{"name":"Default"}}}}}}`
	w0 := httptest.NewRecorder()
	r0 := httptest.NewRequest(http.MethodPost, "/v3/auth/tokens", strings.NewReader(login))
	r0.Host = "127.0.0.1:5000"
	r0.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w0, r0)
	if w0.Code != http.StatusCreated {
		t.Fatal(w0.Body.String())
	}
	tok := w0.Header().Get("X-Subject-Token")

	// region
	wr := httptest.NewRecorder()
	rr := httptest.NewRequest(http.MethodPost, "/v3/regions", strings.NewReader(`{"region":{"id":"R1","description":"r"}}`))
	rr.Host = "127.0.0.1:5000"
	rr.Header.Set("X-Auth-Token", tok)
	rr.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(wr, rr)
	if wr.Code != http.StatusCreated {
		t.Fatalf("create region %d %s", wr.Code, wr.Body.String())
	}

	ws := httptest.NewRecorder()
	rs := httptest.NewRequest(http.MethodPost, "/v3/services", strings.NewReader(`{"service":{"type":"identity","name":"keystone"}}`))
	rs.Host = "127.0.0.1:5000"
	rs.Header.Set("X-Auth-Token", tok)
	rs.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(ws, rs)
	if ws.Code != http.StatusCreated {
		t.Fatalf("create service %d %s", ws.Code, ws.Body.String())
	}
	var svcWrap struct {
		Service struct {
			ID string `json:"id"`
		} `json:"service"`
	}
	if err := json.Unmarshal(ws.Body.Bytes(), &svcWrap); err != nil {
		t.Fatal(err)
	}
	svcID := svcWrap.Service.ID

	we := httptest.NewRecorder()
	re := httptest.NewRequest(http.MethodPost, "/v3/endpoints", strings.NewReader(
		`{"endpoint":{"service_id":"`+svcID+`","region_id":"R1","interface":"public","url":"http://example/v3"}}`,
	))
	re.Host = "127.0.0.1:5000"
	re.Header.Set("X-Auth-Token", tok)
	re.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(we, re)
	if we.Code != http.StatusCreated {
		t.Fatalf("create endpoint %d %s", we.Code, we.Body.String())
	}

	// cannot delete region while endpoint references it
	wd := httptest.NewRecorder()
	rd := httptest.NewRequest(http.MethodDelete, "/v3/regions/R1", nil)
	rd.Host = "127.0.0.1:5000"
	rd.Header.Set("X-Auth-Token", tok)
	r.ServeHTTP(wd, rd)
	if wd.Code != http.StatusConflict {
		t.Fatalf("delete region with refs want 409 got %d %s", wd.Code, wd.Body.String())
	}

	// role assignment POST (idempotent with existing admin assignment)
	body := `{"role_assignment":{"scope":{"project":{"id":"` + fix.ProjectID + `"}},"user":{"id":"` + fix.UserID + `"},"role":{"id":"` + fix.RoleID + `"}}}`
	// admin already has role; use same role idempotent
	wa := httptest.NewRecorder()
	ra := httptest.NewRequest(http.MethodPost, "/v3/role_assignments", strings.NewReader(body))
	ra.Host = "127.0.0.1:5000"
	ra.Header.Set("X-Auth-Token", tok)
	ra.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(wa, ra)
	if wa.Code != http.StatusCreated {
		t.Fatalf("role_assignments POST %d %s", wa.Code, wa.Body.String())
	}

	wl := httptest.NewRecorder()
	rl := httptest.NewRequest(http.MethodGet, "/v3/projects/"+fix.ProjectID+"/users/"+fix.UserID+"/roles", nil)
	rl.Host = "127.0.0.1:5000"
	rl.Header.Set("X-Auth-Token", tok)
	r.ServeHTTP(wl, rl)
	if wl.Code != http.StatusOK {
		t.Fatalf("list project user roles %d %s", wl.Code, wl.Body.String())
	}
}
