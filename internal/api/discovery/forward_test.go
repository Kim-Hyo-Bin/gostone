package discovery

import (
	"crypto/tls"
	"net/http/httptest"
	"testing"
)

func TestPreferredV3URL_forwardedWhenTrusted(t *testing.T) {
	t.Cleanup(ResetForwardedTrust)
	SetTrustForwardedHeaders(true)
	r := httptest.NewRequest("GET", "/", nil)
	r.Host = "internal:5000"
	r.Header.Set("X-Forwarded-Host", "identity.example.com")
	r.Header.Set("X-Forwarded-Proto", "https")
	if u := PreferredV3URL(r); u != "https://identity.example.com/v3/" {
		t.Fatalf("got %q", u)
	}
}

func TestPreferredV3URL_forwardedIgnoredWhenUntrusted(t *testing.T) {
	t.Cleanup(ResetForwardedTrust)
	r := httptest.NewRequest("GET", "/", nil)
	r.Host = "internal:5000"
	r.Header.Set("X-Forwarded-Host", "evil.com")
	r.Header.Set("X-Forwarded-Proto", "https")
	if u := PreferredV3URL(r); u != "http://internal:5000/v3/" {
		t.Fatalf("got %q", u)
	}
}

func TestPreferredV3URL_forwardedFirstCSV(t *testing.T) {
	t.Cleanup(ResetForwardedTrust)
	SetTrustForwardedHeaders(true)
	r := httptest.NewRequest("GET", "/", nil)
	r.Host = "10.0.0.1:5000"
	r.Header.Set("X-Forwarded-Host", "api.openstack.local, other")
	r.Header.Set("X-Forwarded-Proto", "https, http")
	if u := PreferredV3URL(r); u != "https://api.openstack.local/v3/" {
		t.Fatalf("got %q", u)
	}
}

func TestPreferredV3URL_tlsWhenUntrusted(t *testing.T) {
	t.Cleanup(ResetForwardedTrust)
	r := httptest.NewRequest("GET", "/", nil)
	r.Host = "k:5000"
	r.TLS = &tls.ConnectionState{}
	if u := PreferredV3URL(r); u != "https://k:5000/v3/" {
		t.Fatalf("got %q", u)
	}
}
