package v3

import "github.com/gin-gonic/gin"

func registerV3Auth(v3 *gin.RouterGroup, h *Hub) {
	v3.POST("/ec2tokens", h.postEc2Tokens)
	v3.POST("/s3tokens", h.postS3Tokens)

	v3.POST("/auth/tokens", h.postAuthTokens)
	v3.GET("/auth/tokens", h.getAuthTokens)
	v3.HEAD("/auth/tokens", h.headAuthTokens)
	v3.DELETE("/auth/tokens", h.deleteAuthTokens)

	v3.Any("/auth/tokens/OS-PKI/revoked", h.notImplementedOSPKIRevoked)

	authFed := v3.Group("/auth/OS-FEDERATION")
	authFed.Any("/saml2", h.notImplementedFederationAuth)
	authFed.Any("/saml2/ecp", h.notImplementedFederationAuth)
	authFed.Any("/websso/:protocol_id", h.notImplementedFederationAuth)
	authFed.Any("/identity_providers/:idp_id/protocols/:protocol_id/websso", h.notImplementedFederationAuth)
}
