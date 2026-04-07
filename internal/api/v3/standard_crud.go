package v3

import "github.com/gin-gonic/gin"

func registerV3StandardCRUD(v3 *gin.RouterGroup, h *Hub) {
	standard := []struct {
		collection string
		memberID   string
	}{
		{"groups", "group_id"},
		{"regions", "region_id"},
		{"credentials", "credential_id"},
		{"endpoints", "endpoint_id"},
		{"services", "service_id"},
		{"policies", "policy_id"},
		{"limits", "limit_id"},
		{"registered_limits", "registered_limit_id"},
	}
	for _, r := range standard {
		registerCRUD(v3, r.collection, r.memberID, h)
	}

	registerAny(v3, "/limits/model", "/v3/limits/model", h)

	registerAny(v3, "/ec2tokens", "POST /v3/ec2tokens", h)
	registerAny(v3, "/s3tokens", "POST /v3/s3tokens", h)
}
