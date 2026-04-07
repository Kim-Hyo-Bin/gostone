package handlers

import (
	"github.com/Kim-Hyo-Bin/gostone/internal/ops/discovery"
	"github.com/gin-gonic/gin"
)

func registerDiscovery(r *gin.Engine, h *Hub) {
	r.GET("/", h.getVersionDiscovery)
	r.GET("/v3", h.getV3VersionDocument)
}

func (h *Hub) getVersionDiscovery(c *gin.Context) {
	_ = h
	discovery.ServeRoot(c.Writer, c.Request)
}

func (h *Hub) getV3VersionDocument(c *gin.Context) {
	_ = h
	discovery.ServeV3Summary(c.Writer, c.Request)
}
