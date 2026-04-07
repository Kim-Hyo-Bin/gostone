package v3

import "github.com/gin-gonic/gin"

func registerV3Auth(v3 *gin.RouterGroup, h *Hub) {
	v3.POST("/auth/tokens", h.postAuthTokens)
	v3.GET("/auth/tokens", h.getAuthTokens)
	v3.HEAD("/auth/tokens", h.headAuthTokens)
	v3.DELETE("/auth/tokens", h.deleteAuthTokens)

	registerAny(v3, "/auth/projects", "/v3/auth/projects", h)
	registerAny(v3, "/OS-FEDERATION/projects", "/v3/OS-FEDERATION/projects", h)

	registerAny(v3, "/auth/domains", "/v3/auth/domains", h)
	registerAny(v3, "/OS-FEDERATION/domains", "/v3/OS-FEDERATION/domains", h)

	registerAny(v3, "/auth/system", "/v3/auth/system", h)
	registerAny(v3, "/auth/catalog", "/v3/auth/catalog", h)

	registerAny(v3, "/auth/tokens/OS-PKI/revoked", "/v3/auth/tokens/OS-PKI/revoked", h)

	authFed := v3.Group("/auth/OS-FEDERATION")
	registerAny(authFed, "/saml2", "/v3/auth/OS-FEDERATION/saml2", h)
	registerAny(authFed, "/saml2/ecp", "/v3/auth/OS-FEDERATION/saml2/ecp", h)
	registerAny(authFed, "/websso/:protocol_id", "/v3/auth/OS-FEDERATION/websso/:protocol_id", h)
	registerAny(
		authFed,
		"/identity_providers/:idp_id/protocols/:protocol_id/websso",
		"/v3/auth/OS-FEDERATION/identity_providers/:idp_id/protocols/:protocol_id/websso",
		h,
	)
}
