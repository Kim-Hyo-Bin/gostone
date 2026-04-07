package handlers

import "github.com/gin-gonic/gin"

func registerV3Federation(v3 *gin.RouterGroup, h *Hub) {
	fed := v3.Group("/OS-FEDERATION")

	// Longer identity_provider routes first; use :idp_id everywhere (Gin wildcard rules).
	registerAny(fed, "/saml2/metadata", "/v3/OS-FEDERATION/saml2/metadata", h)
	registerAny(
		fed,
		"/identity_providers/:idp_id/protocols/:protocol_id/auth",
		"/v3/OS-FEDERATION/identity_providers/:idp_id/protocols/:protocol_id/auth",
		h,
	)
	registerAny(
		fed,
		"/identity_providers/:idp_id/protocols/:protocol_id",
		"/v3/OS-FEDERATION/identity_providers/:idp_id/protocols/:protocol_id",
		h,
	)
	registerAny(
		fed,
		"/identity_providers/:idp_id/protocols",
		"/v3/OS-FEDERATION/identity_providers/:idp_id/protocols",
		h,
	)

	registerCRUD(fed, "identity_providers", "idp_id", h)
	registerCRUD(fed, "mappings", "mapping_id", h)
	registerCRUD(fed, "service_providers", "service_provider_id", h)
}
