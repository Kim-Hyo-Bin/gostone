//go:build upstream

package integration

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

func upstreamHTTPClient() *http.Client {
	return &http.Client{Timeout: 30 * time.Second}
}

func upstreamBase(t *testing.T) string {
	t.Helper()
	base := strings.TrimSuffix(strings.TrimSpace(os.Getenv("KEYSTONE_UPSTREAM_URL")), "/")
	if base == "" {
		t.Fatal("KEYSTONE_UPSTREAM_URL is required, e.g. http://127.0.0.1:15000")
	}
	return base
}

func waitIdentityV3(t *testing.T, base string) {
	t.Helper()
	c := upstreamHTTPClient()
	deadline := time.Now().Add(2 * time.Minute)
	for time.Now().Before(deadline) {
		resp, err := c.Get(base + "/v3")
		if err == nil {
			b, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK && len(b) > 0 {
				return
			}
		}
		time.Sleep(2 * time.Second)
	}
	t.Fatal("Keystone GET /v3 did not return 200 in time; is the upstream stack running?")
}

func adminPassword() string {
	if v := os.Getenv("KEYSTONE_ADMIN_PASSWORD"); v != "" {
		return v
	}
	return "admin"
}

// TestUpstreamKeystoneVersionDiscovery checks Identity version discovery (GET /v3).
func TestUpstreamKeystoneVersionDiscovery(t *testing.T) {
	base := upstreamBase(t)
	waitIdentityV3(t, base)

	resp, err := upstreamHTTPClient().Get(base + "/v3")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET /v3: status %d body %s", resp.StatusCode, string(b))
	}
	var doc struct {
		Version struct {
			ID string `json:"id"`
		} `json:"version"`
	}
	if err := json.Unmarshal(b, &doc); err != nil {
		t.Fatalf("parse /v3 JSON: %v body %s", err, string(b))
	}
	if doc.Version.ID == "" || !strings.HasPrefix(doc.Version.ID, "v3") {
		t.Fatalf("expected version.id to start with v3, got %q", doc.Version.ID)
	}
}

// TestUpstreamKeystoneHTTP exercises password auth (project-scoped) and GET /v3/users.
func TestUpstreamKeystoneHTTP(t *testing.T) {
	base := upstreamBase(t)
	waitIdentityV3(t, base)
	adminPW := adminPassword()

	// Project-scoped: real Keystone rejects identity:list_users with an unscoped token (403).
	body := map[string]any{
		"auth": map[string]any{
			"identity": map[string]any{
				"methods": []string{"password"},
				"password": map[string]any{
					"user": map[string]any{
						"name":     "admin",
						"password": adminPW,
						"domain":   map[string]any{"name": "Default"},
					},
				},
			},
			"scope": map[string]any{
				"project": map[string]any{
					"name":   "admin",
					"domain": map[string]any{"name": "Default"},
				},
			},
		},
	}
	raw, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest(http.MethodPost, base+"/v3/auth/tokens", bytes.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := upstreamHTTPClient().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("auth status %d: %s", resp.StatusCode, string(b))
	}
	token := resp.Header.Get("X-Subject-Token")
	if token == "" {
		t.Fatal("missing X-Subject-Token")
	}

	req2, err := http.NewRequest(http.MethodGet, base+"/v3/users", nil)
	if err != nil {
		t.Fatal(err)
	}
	req2.Header.Set("X-Auth-Token", token)
	resp2, err := upstreamHTTPClient().Do(req2)
	if err != nil {
		t.Fatal(err)
	}
	defer resp2.Body.Close()
	b, err := io.ReadAll(resp2.Body)
	if err != nil {
		t.Fatal(err)
	}
	if resp2.StatusCode != http.StatusOK {
		t.Fatalf("users status %d: %s", resp2.StatusCode, string(b))
	}
	var wrap struct {
		Users []struct {
			Name string `json:"name"`
		} `json:"users"`
	}
	if err := json.Unmarshal(b, &wrap); err != nil {
		t.Fatal(err)
	}
	found := false
	for _, u := range wrap.Users {
		if u.Name == "admin" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected admin in users: %s", string(b))
	}
}

// TestUpstreamKeystoneUnscopedListUsersForbidden documents default upstream policy:
// unscoped tokens cannot list users (contrast with gostone’s permissive integration path).
func TestUpstreamKeystoneUnscopedListUsersForbidden(t *testing.T) {
	base := upstreamBase(t)
	waitIdentityV3(t, base)
	adminPW := adminPassword()

	body := map[string]any{
		"auth": map[string]any{
			"identity": map[string]any{
				"methods": []string{"password"},
				"password": map[string]any{
					"user": map[string]any{
						"name":     "admin",
						"password": adminPW,
						"domain":   map[string]any{"name": "Default"},
					},
				},
			},
		},
	}
	raw, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest(http.MethodPost, base+"/v3/auth/tokens", bytes.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := upstreamHTTPClient().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("unscoped auth status %d: %s", resp.StatusCode, string(b))
	}
	token := resp.Header.Get("X-Subject-Token")
	if token == "" {
		t.Fatal("missing X-Subject-Token")
	}

	req2, err := http.NewRequest(http.MethodGet, base+"/v3/users", nil)
	if err != nil {
		t.Fatal(err)
	}
	req2.Header.Set("X-Auth-Token", token)
	resp2, err := upstreamHTTPClient().Do(req2)
	if err != nil {
		t.Fatal(err)
	}
	defer resp2.Body.Close()
	b, err := io.ReadAll(resp2.Body)
	if err != nil {
		t.Fatal(err)
	}
	if resp2.StatusCode != http.StatusForbidden {
		t.Fatalf("expected 403 Forbidden for identity:list_users with unscoped token, got %d: %s", resp2.StatusCode, string(b))
	}
}
