package v3

import (
	"github.com/Kim-Hyo-Bin/gostone/internal/auth"
	"github.com/Kim-Hyo-Bin/gostone/internal/common/httperr"
	"github.com/gin-gonic/gin"
)

// requireAuthPolicy writes 401/403 and returns false if the caller may not perform action.
func (h *Hub) requireAuthPolicy(c *gin.Context, action string, target map[string]string) (auth.Context, bool) {
	actx, ok := auth.FromGin(c)
	if !ok {
		httperr.Unauthorized(c, "Unauthorized")
		return auth.Context{}, false
	}
	if !h.Policy.Allow(action, actx, target) {
		httperr.Forbidden(c, "You are not authorized to perform the requested action.")
		return auth.Context{}, false
	}
	return actx, true
}
