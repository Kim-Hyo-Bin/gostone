package v3

import "github.com/gin-gonic/gin"

func registerV3RolesAndAssignments(v3 *gin.RouterGroup, h *Hub) {
	// Register before /roles/:role_id CRUD — Gin requires the same wildcard name
	// for all routes under /roles/:...
	v3.Any("/roles/:role_id/implies", h.notImplementedRoleInference)
	registerV3RolesAPI(v3, h)
	v3.Any("/role_inferences", h.notImplementedRoleInference)
	registerV3RoleAssignments(v3, h)
}
