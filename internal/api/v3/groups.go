package v3

import "github.com/gin-gonic/gin"

func registerV3Groups(v3 *gin.RouterGroup, h *Hub) {
	registerAny(v3, "/groups/:group_id/users", "/v3/groups/:group_id/users", h)
	registerAny(v3, "/groups/:group_id/users/:user_id", "/v3/groups/:group_id/users/:user_id", h)
}
