// Package discovery implements version discovery responses for the Identity API.
package discovery

import (
	"encoding/json"
	"net/http"
)

const mediaTypeJSON = "application/vnd.openstack.identity-v3+json"

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
	return versionDoc{
		ID:      "v3.14",
		Status:  "stable",
		Updated: "2020-04-07T00:00:00Z",
		Links: []link{
			{Rel: "self", Href: u},
		},
		MediaTypes: []media{
			{Base: "application/json", Type: mediaTypeJSON},
		},
	}
}

// PreferredV3URL builds the canonical Identity v3 base URL for this request.
func PreferredV3URL(r *http.Request) string {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	return scheme + "://" + r.Host + "/v3/"
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
