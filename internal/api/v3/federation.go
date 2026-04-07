package v3

import "github.com/gin-gonic/gin"

func registerV3Federation(v3 *gin.RouterGroup, h *Hub) {
	fed := v3.Group("/OS-FEDERATION")
	fed.GET("/saml2/metadata", h.getSAML2Metadata)
	fed.GET("/identity_providers/:idp_id/protocols/:protocol_id/auth", h.fedAuthPlaceholder)
	fed.POST("/identity_providers/:idp_id/protocols/:protocol_id/auth", h.fedAuthPlaceholder)
	registerV3FederationCRUD(fed, h)
}
