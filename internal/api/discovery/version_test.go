package discovery

import (
	"crypto/tls"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPreferredV3URL(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Host = "keystone:5000"
	r.TLS = nil
	if u := PreferredV3URL(r); u != "http://keystone:5000/v3/" {
		t.Fatalf("http: %q", u)
	}
	r.TLS = &tls.ConnectionState{}
	if u := PreferredV3URL(r); u != "https://keystone:5000/v3/" {
		t.Fatalf("https: %q", u)
	}
}

func TestServeV3Summary(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/v3", nil)
	r.Host = "127.0.0.1:5000"
	ServeV3Summary(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("code %d", w.Code)
	}
	var body struct {
		Version struct {
			ID     string `json:"id"`
			Status string `json:"status"`
		} `json:"version"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Version.ID != "v3.14" || body.Version.Status != "stable" {
		t.Fatalf("%+v", body)
	}
}

func TestServeRoot(t *testing.T) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Host = "example.org"
	ServeRoot(w, r)
	if w.Code != http.StatusMultipleChoices {
		t.Fatalf("code %d", w.Code)
	}
	loc := w.Header().Get("Location")
	if loc != "http://example.org/v3/" {
		t.Fatalf("Location %q", loc)
	}
}
