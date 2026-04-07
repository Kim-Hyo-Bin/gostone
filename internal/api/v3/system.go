package v3

import "github.com/gin-gonic/gin"

func registerV3System(v3 *gin.RouterGroup, h *Hub) {
	registerAny(v3, "/system/users/:user_id/roles", "/v3/system/users/:user_id/roles", h)
	registerAny(v3, "/system/users/:user_id/roles/:role_id", "/v3/system/users/:user_id/roles/:role_id", h)
	registerAny(v3, "/system/groups/:group_id/roles", "/v3/system/groups/:group_id/roles", h)
	registerAny(v3, "/system/groups/:group_id/roles/:role_id", "/v3/system/groups/:group_id/roles/:role_id", h)
}
