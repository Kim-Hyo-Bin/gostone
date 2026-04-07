package auth

import "strings"

// SkipAuth reports routes that must work without X-Auth-Token (Keystone public auth paths).
func SkipAuth(method, absPath string) bool {
	method = strings.ToUpper(method)
	switch absPath {
	case "/v3/auth/tokens":
		return method == "POST"
	default:
		return false
	}
}
