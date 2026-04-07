package v3

import "github.com/gin-gonic/gin"

func registerV3Trust(v3 *gin.RouterGroup, h *Hub) {
	trust := v3.Group("/OS-TRUST")
	registerCRUD(trust, "trusts", "trust_id", h)
	registerAny(trust, "/trusts/:trust_id/roles", "/v3/OS-TRUST/trusts/:trust_id/roles", h)
	registerAny(trust, "/trusts/:trust_id/roles/:role_id", "/v3/OS-TRUST/trusts/:trust_id/roles/:role_id", h)
}
