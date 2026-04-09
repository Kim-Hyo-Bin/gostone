package conf

import (
	"fmt"
	"strings"
)

// ListenBinding is one HTTP bind address for gostone (shared handler stack).
type ListenBinding struct {
	// Name is listen (single-interface mode) or public | admin | internal.
	Name string
	Addr string
}

// ListenBindings derives bind addresses from [service].
//
// Single-interface mode (default): none of listen_public, listen_admin, listen_internal are set.
// Then one server uses listen (default :5000).
//
// Multi-interface mode: at least one of listen_public, listen_admin, listen_internal is non-empty.
// The same HTTP handler is bound on each address. Public address is listen_public if set, else listen
// if set; if both are empty, no public listener is started (admin- or internal-only is allowed).
func ListenBindings(s *Service) ([]ListenBinding, error) {
	if s == nil {
		return nil, fmt.Errorf("nil service config")
	}
	pub := strings.TrimSpace(s.ListenPublic)
	adm := strings.TrimSpace(s.ListenAdmin)
	intl := strings.TrimSpace(s.ListenInternal)
	single := strings.TrimSpace(s.Listen)

	multi := pub != "" || adm != "" || intl != ""
	if !multi {
		if single == "" {
			single = ":5000"
		}
		return []ListenBinding{{Name: "listen", Addr: single}}, nil
	}

	var out []ListenBinding
	publicAddr := pub
	if publicAddr == "" {
		publicAddr = single
	}
	if publicAddr != "" {
		out = append(out, ListenBinding{Name: "public", Addr: publicAddr})
	}
	if adm != "" {
		out = append(out, ListenBinding{Name: "admin", Addr: adm})
	}
	if intl != "" {
		out = append(out, ListenBinding{Name: "internal", Addr: intl})
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("service: set listen_public or listen for a public bind, or at least one of listen_admin / listen_internal")
	}
	return out, nil
}
