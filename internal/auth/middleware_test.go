package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Kim-Hyo-Bin/gostone/internal/token"
	"github.com/gin-gonic/gin"
)

func TestMiddleware_skipAuthPath(t *testing.T) {
	j := &token.JWT{Secret: []byte("s"), Issuer: "i", TTL: time.Hour}
	r := gin.New()
	r.Use(Middleware(j))
	r.POST("/v3/auth/tokens", func(c *gin.Context) { c.Status(http.StatusTeapot) })

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v3/auth/tokens", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusTeapot {
		t.Fatalf("code %d", w.Code)
	}
}

func TestMiddleware_missingToken(t *testing.T) {
	j := &token.JWT{Secret: []byte("s"), Issuer: "i", TTL: time.Hour}
	r := gin.New()
	r.Use(Middleware(j))
	r.GET("/v3/users", func(c *gin.Context) { c.Status(http.StatusOK) })

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v3/users", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("code %d body %s", w.Code, w.Body.String())
	}
}

func TestMiddleware_invalidToken(t *testing.T) {
	j := &token.JWT{Secret: []byte("s"), Issuer: "i", TTL: time.Hour}
	r := gin.New()
	r.Use(Middleware(j))
	r.GET("/v3/users", func(c *gin.Context) { c.Status(http.StatusOK) })

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/v3/users", nil)
	req.Header.Set("X-Auth-Token", "bad")
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("code %d", w.Code)
	}
}

func TestMiddleware_validToken_setsContext(t *testing.T) {
	j := &token.JWT{Secret: []byte("secret-key-middleware-test"), Issuer: "i", TTL: time.Hour}
	tok, _, err := j.Issue("uid", "did", "pid", []string{"admin"})
	if err != nil {
		t.Fatal(err)
	}

	var got Context
	r := gin.New()
	r.Use(Middleware(j))
	r.GET("/x", func(c *gin.Context) {
		var ok bool
		got, ok = FromGin(c)
		if !ok {
			t.Error("no context")
		}
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("X-Auth-Token", tok)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK || got.UserID != "uid" {
		t.Fatalf("code=%d ctx=%+v", w.Code, got)
	}
	if w.Header().Get("Vary") != "X-Auth-Token" {
		t.Fatalf("Vary: %q", w.Header().Get("Vary"))
	}
}
