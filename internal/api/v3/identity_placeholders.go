package v3

import (
	"github.com/Kim-Hyo-Bin/gostone/internal/common/httperr"
	"github.com/gin-gonic/gin"
)

// Explicit 501 handlers (Keystone-compatible paths) so clients get stable messages instead of generic stub labels.

func (h *Hub) notImplementedOSPKIRevoked(c *gin.Context) {
	httperr.NotImplemented(c, "OS-PKI revoked token list (GET /v3/auth/tokens/OS-PKI/revoked) is not implemented.")
}

func (h *Hub) notImplementedFederationAuth(c *gin.Context) {
	httperr.NotImplemented(c, "OS-FEDERATION protocol auth (SAML2, ECP, WebSSO) is not implemented; use POST /v3/auth/tokens with password, token, or application_credential.")
}

func (h *Hub) notImplementedRoleInference(c *gin.Context) {
	httperr.NotImplemented(c, "Implied roles and role inference (/v3/roles/:id/implies, /v3/role_inferences) are not implemented.")
}
