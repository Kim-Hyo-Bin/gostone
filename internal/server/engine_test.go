package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Kim-Hyo-Bin/gostone/internal/listenctx"
)

func TestNewEngine_openstackRequestID(t *testing.T) {
	eng := NewEngine(newTestHub(t), EngineOptions{})
	h := listenctx.WrapHandler("public", eng)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req.Host = "127.0.0.1:5000"
	h.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("code %d", w.Code)
	}
	if w.Header().Get("X-OpenStack-Request-Id") == "" {
		t.Fatal("missing X-OpenStack-Request-Id")
	}
}

func TestNewEngine_preservesIncomingRequestID(t *testing.T) {
	eng := NewEngine(newTestHub(t), EngineOptions{})
	h := listenctx.WrapHandler("public", eng)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req.Host = "127.0.0.1:5000"
	req.Header.Set("X-OpenStack-Request-Id", "req-abc")
	h.ServeHTTP(w, req)
	if got := w.Header().Get("X-OpenStack-Request-Id"); got != "req-abc" {
		t.Fatalf("got %q", got)
	}
}

func TestNewEngine_passwordTotpComposite_errorIncludesRequestID(t *testing.T) {
	eng := NewEngine(newTestHub(t), EngineOptions{})
	body := `{"auth":{"identity":{"methods":["password","totp"],"password":{"user":{"name":"admin","password":"x","domain":{"name":"Default"}}}}}}`
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v3/auth/tokens", strings.NewReader(body))
	req.Host = "127.0.0.1:5000"
	req.Header.Set("Content-Type", "application/json")
	eng.ServeHTTP(w, req)
	if w.Code != http.StatusNotImplemented {
		t.Fatalf("code %d %s", w.Code, w.Body.String())
	}
	var wrap struct {
		Error struct {
			Code      int    `json:"code"`
			RequestID string `json:"request_id"`
		} `json:"error"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &wrap); err != nil {
		t.Fatal(err)
	}
	if wrap.Error.RequestID == "" {
		t.Fatal("expected error.request_id")
	}
}
