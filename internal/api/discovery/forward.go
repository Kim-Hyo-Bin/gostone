package discovery

import (
	"net/http"
	"strings"
	"sync/atomic"
)

var trustForwarded atomic.Bool

// SetTrustForwardedHeaders enables use of X-Forwarded-Host and X-Forwarded-Proto for discovery
// and other request-derived URLs (e.g. WWW-Authenticate). Only enable behind a trusted reverse proxy.
func SetTrustForwardedHeaders(v bool) {
	trustForwarded.Store(v)
}

// TrustForwardedHeaders reports whether forwarded headers are honored.
func TrustForwardedHeaders() bool {
	return trustForwarded.Load()
}

// ResetForwardedTrust turns off forwarded-header trust (for tests).
func ResetForwardedTrust() {
	trustForwarded.Store(false)
}

func forwardedHost(r *http.Request) string {
	if trustForwarded.Load() {
		if h := r.Header.Get("X-Forwarded-Host"); h != "" {
			return firstCSVToken(h)
		}
	}
	return r.Host
}

func forwardedScheme(r *http.Request) string {
	if trustForwarded.Load() {
		if p := r.Header.Get("X-Forwarded-Proto"); p != "" {
			s := strings.ToLower(strings.TrimSpace(firstCSVToken(p)))
			if s == "http" || s == "https" {
				return s
			}
		}
	}
	if r.TLS != nil {
		return "https"
	}
	return "http"
}

func firstCSVToken(s string) string {
	s = strings.TrimSpace(s)
	if i := strings.IndexByte(s, ','); i >= 0 {
		s = s[:i]
	}
	return strings.TrimSpace(s)
}
