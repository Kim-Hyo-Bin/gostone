package v3

import "github.com/gin-gonic/gin"

// registerV3 mounts Identity API v3 routes aligned with OpenStack Keystone stable/2024.2.
// Registration order matters for Gin (wildcard name consistency and path specificity).
func registerV3(v3 *gin.RouterGroup, h *Hub) {
	registerV3DomainsAPI(v3, h)
	registerV3ProjectsAPI(v3, h)
	registerV3CatalogAPI(v3, h)
	registerV3RolesAndAssignments(v3, h)
	registerV3Users(v3, h)
	registerV3GroupsAPI(v3, h)
	registerV3PoliciesBundle(v3, h)
	registerV3CredentialsAPI(v3, h)
	registerV3UserEC2Credentials(v3, h)
	registerV3Auth(v3, h)
	registerV3AuthResources(v3, h)
	registerV3ProjectTags(v3, h)
	registerV3UserMisc(v3, h)
	registerV3UserAppCredentials(v3, h)
	registerV3OSRevokeOAuthCert(v3, h)
	registerV3OSEPFilter(v3, h)
	registerV3InheritAndSystem(v3, h)
	registerV3Trust(v3, h)
	registerV3Federation(v3, h)
}
