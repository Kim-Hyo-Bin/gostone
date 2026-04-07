package v3

import (
	"github.com/Kim-Hyo-Bin/gostone/internal/auth"
	"github.com/gin-gonic/gin"
)

// Mount registers Identity v3 routes under /v3 with auth middleware.
func Mount(r *gin.Engine, h *Hub) {
	v3 := r.Group("/v3")
	v3.Use(auth.Middleware(h.Tokens))
	registerV3(v3, h)
}
