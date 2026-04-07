package v3

import "github.com/gin-gonic/gin"

// registerV3 mounts Identity API v3 routes aligned with OpenStack Keystone stable/2024.2.
// Registration order matters for Gin (wildcard name consistency and path specificity).
func registerV3(v3 *gin.RouterGroup, h *Hub) {
	registerV3DomainsAPI(v3, h)
	registerV3ProjectsAPI(v3, h)
	registerV3RolesAndAssignments(v3, h)
	registerV3Users(v3, h)
	registerV3StandardCRUD(v3, h)
	registerV3Auth(v3, h)
	registerV3ProjectsAndUsers(v3, h)
	registerV3Domains(v3, h)
	registerV3Groups(v3, h)
	registerV3EndpointsPolicy(v3, h)
	registerV3Trust(v3, h)
	registerV3Federation(v3, h)
	registerV3Inherit(v3, h)
	registerV3System(v3, h)
	registerV3OSExtensions(v3, h)
}
