package v3

import "github.com/gin-gonic/gin"

func registerV3EndpointsPolicy(v3 *gin.RouterGroup, h *Hub) {
	registerAny(v3, "/endpoints/:endpoint_id/OS-ENDPOINT-POLICY/policy", "/v3/endpoints/:endpoint_id/OS-ENDPOINT-POLICY/policy", h)

	registerAny(v3, "/policies/:policy_id/OS-ENDPOINT-POLICY/endpoints", "/v3/policies/:policy_id/OS-ENDPOINT-POLICY/endpoints", h)
	registerAny(v3, "/policies/:policy_id/OS-ENDPOINT-POLICY/endpoints/:endpoint_id", "/v3/policies/:policy_id/OS-ENDPOINT-POLICY/endpoints/:endpoint_id", h)
	registerAny(v3, "/policies/:policy_id/OS-ENDPOINT-POLICY/services/:service_id", "/v3/policies/:policy_id/OS-ENDPOINT-POLICY/services/:service_id", h)
	registerAny(v3, "/policies/:policy_id/OS-ENDPOINT-POLICY/services/:service_id/regions/:region_id", "/v3/policies/:policy_id/OS-ENDPOINT-POLICY/services/:service_id/regions/:region_id", h)
}
