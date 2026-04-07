package auth

import "strings"

// SkipAuth reports routes that must work without X-Auth-Token (Keystone public auth paths).
func SkipAuth(method, absPath string) bool {
	method = strings.ToUpper(method)
	switch absPath {
	case "/v3/auth/tokens", "/v3/ec2tokens", "/v3/s3tokens":
		return method == "POST"
	case "/v3/OS-OAUTH1/request_token", "/v3/OS-OAUTH1/access_token":
		return method == "POST"
	default:
		return false
	}
}
