package server

import (
	"github.com/Kim-Hyo-Bin/gostone/internal/api/v3"
	"github.com/gin-gonic/gin"
)

// EngineOptions configures optional HTTP middleware (shared across all listeners).
type EngineOptions struct {
	EnforceAdminOnly  bool
	AdminOnlyPrefixes []string
	JSONAccessLogs    bool
}

// NewEngine builds one Gin engine with shared middleware and Identity routes.
// The same engine can be served on multiple listen addresses (public / admin / internal), matching
// common Keystone deployments where the WSGI app is identical on each socket.
// Wrap the engine with listenctx.WrapHandler per listener so middleware can tell public from admin.
func NewEngine(hub *v3.Hub, opts EngineOptions) *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery(), requestIDMiddleware(), injectListenBindingMiddleware())
	if opts.JSONAccessLogs {
		r.Use(gin.LoggerWithFormatter(jsonAccessLogFormatter))
	} else {
		r.Use(gin.LoggerWithFormatter(gostoneLogFormatter))
	}
	if opts.EnforceAdminOnly && len(opts.AdminOnlyPrefixes) > 0 {
		r.Use(adminOnlyListenerMiddleware(opts.AdminOnlyPrefixes))
	}
	Register(r, hub)
	return r
}
