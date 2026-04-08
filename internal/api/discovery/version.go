// Package discovery implements version discovery responses for the Identity API.
package discovery

import (
	"encoding/json"
	"net/http"
	"sync"
)

const mediaTypeJSON = "application/vnd.openstack.identity-v3+json"

// DocConfig overrides GET / and GET /v3 version document fields (Keystone-style discovery).
type DocConfig struct {
	VersionID string // e.g. v3.14
	Updated   string // RFC3339
	Status    string // e.g. stable
}

var (
	docMu  sync.RWMutex
	docCfg = DocConfig{
		VersionID: "v3.14",
		Updated:   "2020-04-07T00:00:00Z",
		Status:    "stable",
	}
)

// ConfigureDiscovery sets the advertised Identity API version document (safe to call before serving).
func ConfigureDiscovery(c DocConfig) {
	docMu.Lock()
	defer docMu.Unlock()
	if c.VersionID != "" {
		docCfg.VersionID = c.VersionID
	}
	if c.Updated != "" {
		docCfg.Updated = c.Updated
	}
	if c.Status != "" {
		docCfg.Status = c.Status
	}
}

// ResetDiscoveryDoc restores default discovery document values (for tests).
func ResetDiscoveryDoc() {
	docMu.Lock()
	defer docMu.Unlock()
	docCfg = DocConfig{
		VersionID: "v3.14",
		Updated:   "2020-04-07T00:00:00Z",
		Status:    "stable",
	}
}

func currentDoc() DocConfig {
	docMu.RLock()
	defer docMu.RUnlock()
	return docCfg
}

type versionDoc struct {
	ID         string  `json:"id"`
	Status     string  `json:"status"`
	Updated    string  `json:"updated"`
	Links      []link  `json:"links"`
	MediaTypes []media `json:"media-types"`
}

type link struct {
	Rel  string `json:"rel"`
	Href string `json:"href"`
}

type media struct {
	Base string `json:"base"`
	Type string `json:"type"`
}

func identityV3Doc(baseURL string) versionDoc {
	u := baseURL
	if u != "" && u[len(u)-1] != '/' {
		u += "/"
	}
	d := currentDoc()
	return versionDoc{
		ID:      d.VersionID,
		Status:  d.Status,
		Updated: d.Updated,
		Links: []link{
			{Rel: "self", Href: u},
		},
		MediaTypes: []media{
			{Base: "application/json", Type: mediaTypeJSON},
		},
	}
}

// PreferredV3URL builds the canonical Identity v3 base URL for this request.
// When SetTrustForwardedHeaders(true), uses X-Forwarded-Host and X-Forwarded-Proto when present.
func PreferredV3URL(r *http.Request) string {
	return forwardedScheme(r) + "://" + forwardedHost(r) + "/v3/"
}

// ServeRoot handles GET / (version discovery). Matches Keystone: 300 + versions payload + Location.
func ServeRoot(w http.ResponseWriter, r *http.Request) {
	loc := PreferredV3URL(r)
	v := identityV3Doc(loc)
	body := map[string]any{
		"versions": map[string]any{
			"values": []versionDoc{v},
		},
	}
	raw, err := json.Marshal(body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Location", loc)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusMultipleChoices)
	_, _ = w.Write(raw)
}

// ServeV3Summary handles GET /v3 (single version document).
func ServeV3Summary(w http.ResponseWriter, r *http.Request) {
	loc := PreferredV3URL(r)
	v := identityV3Doc(loc)
	raw, err := json.Marshal(map[string]versionDoc{"version": v})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(raw)
}
