package v3

import "github.com/gin-gonic/gin"

func registerV3RolesAndAssignments(v3 *gin.RouterGroup, h *Hub) {
	// Register before /roles/:role_id CRUD — Gin requires the same wildcard name
	// for all routes under /roles/:...
	registerAny(v3, "/roles/:role_id/implies", "/v3/roles/:role_id/implies", h)
	registerV3RolesAPI(v3, h)
	registerAny(v3, "/role_inferences", "/v3/role_inferences", h)
	registerV3RoleAssignments(v3, h)
}
