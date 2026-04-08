package server

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Kim-Hyo-Bin/gostone/internal/api/v3"
	"github.com/Kim-Hyo-Bin/gostone/internal/listenctx"
	"github.com/Kim-Hyo-Bin/gostone/internal/policy"
	"github.com/Kim-Hyo-Bin/gostone/internal/testutil"
	"github.com/Kim-Hyo-Bin/gostone/internal/token"
	"github.com/gin-gonic/gin"
	"time"
)

func TestAdminOnlyListenerMiddleware_blocksPublic(t *testing.T) {
	gin.SetMode(gin.TestMode)
	gdb, err := testutil.OpenMemory()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := testutil.SeedAdmin(gdb, "x"); err != nil {
		t.Fatal(err)
	}
	mgr, err := token.NewManager(gdb, token.ProviderJWT, "adm-listen-test", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	hub := &v3.Hub{DB: gdb, Tokens: mgr, Policy: policy.Default()}
	eng := NewEngine(hub, EngineOptions{
		EnforceAdminOnly:  true,
		AdminOnlyPrefixes: []string{"/v3/domains"},
	})
	h := listenctx.WrapHandler("public", eng)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v3/domains", nil)
	req.Host = "127.0.0.1:5000"
	h.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("code %d", w.Code)
	}
}

func TestAdminOnlyListenerMiddleware_allowsAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	gdb, err := testutil.OpenMemory()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := testutil.SeedAdmin(gdb, "x"); err != nil {
		t.Fatal(err)
	}
	mgr, err := token.NewManager(gdb, token.ProviderJWT, "adm-listen-test2", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	hub := &v3.Hub{DB: gdb, Tokens: mgr, Policy: policy.Default()}
	eng := NewEngine(hub, EngineOptions{
		EnforceAdminOnly:  true,
		AdminOnlyPrefixes: []string{"/v3/domains"},
	})
	h := listenctx.WrapHandler("admin", eng)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v3/domains", nil)
	req.Host = "127.0.0.1:35357"
	h.ServeHTTP(w, req)
	if w.Code == http.StatusForbidden {
		t.Fatalf("unexpected 403")
	}
}
