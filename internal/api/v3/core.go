package v3

import (
	"github.com/Kim-Hyo-Bin/gostone/internal/common/httperr"
	"github.com/gin-gonic/gin"
)

func (h *Hub) stubRoute(label string) gin.HandlerFunc {
	return func(c *gin.Context) {
		httperr.NotImplemented(c, label)
	}
}

// registerCRUD mounts list/create and member read/update/delete patterns used by
// Keystone flask-restful resources (GET/POST collection, GET/HEAD/PATCH/DELETE member).
func registerCRUD(g *gin.RouterGroup, collection, memberIDParam string, h *Hub) {
	base := "/" + collection
	member := base + "/:" + memberIDParam
	prefix := g.BasePath() // may be empty in Gin <1.9; used only for messages

	g.GET(base, h.stubRoute("GET "+prefix+base))
	g.POST(base, h.stubRoute("POST "+prefix+base))
	g.GET(member, h.stubRoute("GET "+prefix+member))
	g.HEAD(member, h.stubRoute("HEAD "+prefix+member))
	g.PATCH(member, h.stubRoute("PATCH "+prefix+member))
	g.PUT(member, h.stubRoute("PUT "+prefix+member))
	g.DELETE(member, h.stubRoute("DELETE "+prefix+member))
}

func registerAny(g *gin.RouterGroup, relativePath, label string, h *Hub) {
	g.Any(relativePath, h.stubRoute(label))
}
