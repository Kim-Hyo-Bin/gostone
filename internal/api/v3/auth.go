package v3

import "github.com/gin-gonic/gin"

func registerV3Auth(v3 *gin.RouterGroup, h *Hub) {
	v3.POST("/ec2tokens", h.postEc2Tokens)
	v3.POST("/s3tokens", h.postS3Tokens)

	v3.POST("/auth/tokens", h.postAuthTokens)
	v3.GET("/auth/tokens", h.getAuthTokens)
	v3.HEAD("/auth/tokens", h.headAuthTokens)
	v3.DELETE("/auth/tokens", h.deleteAuthTokens)

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
