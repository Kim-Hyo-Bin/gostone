package v3

import "github.com/gin-gonic/gin"

func registerV3OSExtensions(v3 *gin.RouterGroup, h *Hub) {
	revoke := v3.Group("/OS-REVOKE")
	registerAny(revoke, "/events", "/v3/OS-REVOKE/events", h)

	oauth2 := v3.Group("/OS-OAUTH2")
	registerAny(oauth2, "/token", "/v3/OS-OAUTH2/token", h)

	oauth1 := v3.Group("/OS-OAUTH1")
	registerCRUD(oauth1, "consumers", "consumer_id", h)
	registerAny(oauth1, "/request_token", "/v3/OS-OAUTH1/request_token", h)
	registerAny(oauth1, "/access_token", "/v3/OS-OAUTH1/access_token", h)
	registerAny(oauth1, "/authorize/:request_token_id", "/v3/OS-OAUTH1/authorize/:request_token_id", h)

	cert := v3.Group("/OS-SIMPLE-CERT")
	registerAny(cert, "/ca", "/v3/OS-SIMPLE-CERT/ca", h)
	registerAny(cert, "/certificates", "/v3/OS-SIMPLE-CERT/certificates", h)

	epf := v3.Group("/OS-EP-FILTER")
	registerAny(epf, "/endpoints/:endpoint_id/projects", "/v3/OS-EP-FILTER/endpoints/:endpoint_id/projects", h)
	registerAny(epf, "/projects/:project_id/endpoints/:endpoint_id", "/v3/OS-EP-FILTER/projects/:project_id/endpoints/:endpoint_id", h)
	registerAny(epf, "/projects/:project_id/endpoints", "/v3/OS-EP-FILTER/projects/:project_id/endpoints", h)
	registerAny(epf, "/projects/:project_id/endpoint_groups", "/v3/OS-EP-FILTER/projects/:project_id/endpoint_groups", h)
	registerAny(epf, "/endpoint_groups/:endpoint_group_id/endpoints", "/v3/OS-EP-FILTER/endpoint_groups/:endpoint_group_id/endpoints", h)
	registerAny(epf, "/endpoint_groups/:endpoint_group_id/projects", "/v3/OS-EP-FILTER/endpoint_groups/:endpoint_group_id/projects", h)
}
