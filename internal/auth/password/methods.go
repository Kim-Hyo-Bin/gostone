package password

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Kim-Hyo-Bin/gostone/internal/token"
	"gorm.io/gorm"
)

func normalizedAuthMethods(req *PasswordAuthRequest) ([]string, error) {
	if req == nil {
		return nil, errors.New("empty auth request")
	}
	var out []string
	for _, m := range req.Auth.Identity.Methods {
		x := strings.ToLower(strings.TrimSpace(m))
		if x != "" {
			out = append(out, x)
		}
	}
	if len(out) == 0 {
		return nil, errors.New("auth methods required")
	}
	return out, nil
}

func isAuthMethodNotImplemented(m string) bool {
	switch m {
	case "totp", "oauth1", "oauth2", "openid", "mapped", "kerberos", "cert":
		return true
	default:
		return false
	}
}

// issueMultiMethodAuth rejects Keystone-style composite auth (e.g. password+totp) with a clear error.
func issueMultiMethodAuth(_ *gorm.DB, _ *token.Manager, _ *PasswordAuthRequest, methods []string) (string, time.Time, map[string]any, error) {
	for _, m := range methods {
		switch m {
		case "password", "token", "application_credential":
			continue
		case "totp", "oauth1", "oauth2", "openid", "mapped", "kerberos", "cert":
			return "", time.Time{}, nil, fmt.Errorf("authentication method %q is not implemented", m)
		default:
			return "", time.Time{}, nil, fmt.Errorf("unsupported auth method %q in combination", m)
		}
	}
	return "", time.Time{}, nil, fmt.Errorf("unsupported auth method combination: %v (composite MFA is not implemented; use a single method)", methods)
}
