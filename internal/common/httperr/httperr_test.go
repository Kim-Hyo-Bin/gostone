package httperr

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Kim-Hyo-Bin/gostone/internal/api/discovery"
	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestKeystoneURI_HTTP(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/x", nil)
	c.Request.Host = "example.com"

	if u := keystoneURI(c); u != "http://example.com/v3/" {
		t.Fatalf("got %q", u)
	}
}

func TestResponses(t *testing.T) {
	tests := []struct {
		name string
		fn   func(*gin.Context, string)
		code int
	}{
		{"Unauthorized", func(c *gin.Context, m string) { Unauthorized(c, m) }, http.StatusUnauthorized},
		{"BadRequest", func(c *gin.Context, m string) { BadRequest(c, m) }, http.StatusBadRequest},
		{"Forbidden", func(c *gin.Context, m string) { Forbidden(c, m) }, http.StatusForbidden},
		{"NotImplemented", func(c *gin.Context, m string) { NotImplemented(c, m) }, http.StatusNotImplemented},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
			c.Request.Host = "h.test"
			tc.fn(c, "msg")
			if w.Code != tc.code {
				t.Fatalf("code: %d", w.Code)
			}
			var body struct {
				Error struct {
					Code    int    `json:"code"`
					Message string `json:"message"`
				} `json:"error"`
			}
			if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
				t.Fatal(err)
			}
			if body.Error.Message != "msg" {
				t.Fatalf("message: %q", body.Error.Message)
			}
		})
	}
}

func TestUnauthorized_WWWAuthenticate(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Request.Host = "api.example"
	Unauthorized(c, "nope")
	if !strings.HasPrefix(w.Header().Get("WWW-Authenticate"), `Keystone uri="http://api.example/v3/"`) {
		t.Fatalf("header: %q", w.Header().Get("WWW-Authenticate"))
	}
}

func TestUnauthorized_WWWAuthenticate_forwarded(t *testing.T) {
	t.Cleanup(discovery.ResetForwardedTrust)
	discovery.SetTrustForwardedHeaders(true)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Request.Host = "127.0.0.1:5000"
	c.Request.Header.Set("X-Forwarded-Host", "keystone.public")
	c.Request.Header.Set("X-Forwarded-Proto", "https")
	Unauthorized(c, "nope")
	h := w.Header().Get("WWW-Authenticate")
	if !strings.Contains(h, `Keystone uri="https://keystone.public/v3/"`) {
		t.Fatalf("header: %q", h)
	}
}
