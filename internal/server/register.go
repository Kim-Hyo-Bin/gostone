// Package server wires the HTTP surface (Keystone keystone.server role: app factory + mounts).
package server

import (
	"github.com/Kim-Hyo-Bin/gostone/internal/api/discovery"
	"github.com/Kim-Hyo-Bin/gostone/internal/api/v3"
	"github.com/gin-gonic/gin"
)

// Register attaches health, version discovery, and Identity v3 to the engine.
func Register(r *gin.Engine, hub *v3.Hub) {
	v3.RegisterHealth(r, hub)
	registerVersionDiscovery(r)
	v3.Mount(r, hub)
}

func registerVersionDiscovery(r *gin.Engine) {
	r.GET("/", func(c *gin.Context) {
		discovery.ServeRoot(c.Writer, c.Request)
	})
	r.GET("/v3", func(c *gin.Context) {
		discovery.ServeV3Summary(c.Writer, c.Request)
	})
}
