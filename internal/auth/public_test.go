package auth

import (
	"net/http"
	"testing"
)

func TestSkipAuth(t *testing.T) {
	if !SkipAuth(http.MethodPost, "/v3/auth/tokens") {
		t.Fatal("POST tokens should skip")
	}
	if SkipAuth(http.MethodGet, "/v3/auth/tokens") {
		t.Fatal("GET tokens should not skip")
	}
	if SkipAuth(http.MethodPost, "/v3/users") {
		t.Fatal("users should not skip")
	}
}
