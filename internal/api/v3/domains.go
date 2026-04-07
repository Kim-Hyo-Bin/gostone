package v3

import "github.com/gin-gonic/gin"

func registerV3Domains(v3 *gin.RouterGroup, h *Hub) {
	registerAny(v3, "/domains/:domain_id/config", "/v3/domains/:domain_id/config", h)
	registerAny(v3, "/domains/:domain_id/config/:group", "/v3/domains/:domain_id/config/:group", h)
	registerAny(v3, "/domains/:domain_id/config/:group/:option", "/v3/domains/:domain_id/config/:group/:option", h)
	registerAny(v3, "/domains/config/default", "/v3/domains/config/default", h)
	registerAny(v3, "/domains/config/:group/default", "/v3/domains/config/:group/default", h)
	registerAny(v3, "/domains/config/:group/:option/default", "/v3/domains/config/:group/:option/default", h)

	registerAny(v3, "/domains/:domain_id/users/:user_id/roles", "/v3/domains/:domain_id/users/:user_id/roles", h)
	registerAny(v3, "/domains/:domain_id/users/:user_id/roles/:role_id", "/v3/domains/:domain_id/users/:user_id/roles/:role_id", h)
	registerAny(v3, "/domains/:domain_id/groups/:group_id/roles", "/v3/domains/:domain_id/groups/:group_id/roles", h)
	registerAny(v3, "/domains/:domain_id/groups/:group_id/roles/:role_id", "/v3/domains/:domain_id/groups/:group_id/roles/:role_id", h)
}
