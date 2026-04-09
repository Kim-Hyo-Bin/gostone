package v3

import "github.com/gin-gonic/gin"

func registerV3Trust(v3 *gin.RouterGroup, h *Hub) {
	trust := v3.Group("/OS-TRUST")
	registerV3TrustHandlers(trust, h)
}
