package server

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/Kim-Hyo-Bin/gostone/internal/listenctx"
	"github.com/gin-gonic/gin"
)

const listenBindingGinKey = "listen_binding"

func injectListenBindingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if lb := listenctx.Label(c.Request.Context()); lb != "" {
			c.Set(listenBindingGinKey, lb)
		}
		c.Next()
	}
}

func listenTagFromKeys(keys map[any]any) string {
	if keys == nil {
		return ""
	}
	v, ok := keys[listenBindingGinKey].(string)
	if !ok || v == "" {
		return ""
	}
	return "[listen=" + v + "] "
}

// gostoneLogFormatter matches Gin's default access line with an optional [listen=…] tag after [GIN].
func gostoneLogFormatter(param gin.LogFormatterParams) string {
	listenTag := listenTagFromKeys(param.Keys)

	var statusColor, methodColor, resetColor, latencyColor string
	if param.IsOutputColor() {
		statusColor = param.StatusCodeColor()
		methodColor = param.MethodColor()
		resetColor = param.ResetColor()
		latencyColor = param.LatencyColor()
	}

	switch {
	case param.Latency > time.Minute:
		param.Latency = param.Latency.Truncate(time.Second * 10)
	case param.Latency > time.Second:
		param.Latency = param.Latency.Truncate(time.Millisecond * 10)
	case param.Latency > time.Millisecond:
		param.Latency = param.Latency.Truncate(time.Microsecond * 10)
	}

	return fmt.Sprintf("[GIN] %s%v |%s %3d %s|%s %8v %s| %15s |%s %-7s %s %#v\n%s",
		listenTag,
		param.TimeStamp.Format("2006/01/02 - 15:04:05"),
		statusColor, param.StatusCode, resetColor,
		latencyColor, param.Latency, resetColor,
		param.ClientIP,
		methodColor, param.Method, resetColor,
		param.Path,
		param.ErrorMessage,
	)
}

func jsonAccessLogFormatter(param gin.LogFormatterParams) string {
	m := map[string]any{
		"time":        param.TimeStamp.UTC().Format(time.RFC3339Nano),
		"status":      param.StatusCode,
		"latency_ms":  param.Latency.Milliseconds(),
		"method":      param.Method,
		"path":        param.Path,
		"client_ip":   param.ClientIP,
		"body_bytes":  param.BodySize,
		"error":       param.ErrorMessage,
	}
	if param.Keys != nil {
		if v, ok := param.Keys[requestIDGinKey].(string); ok && v != "" {
			m["request_id"] = v
		}
		if v, ok := param.Keys[listenBindingGinKey].(string); ok && v != "" {
			m["listen"] = v
		}
	}
	b, _ := json.Marshal(m)
	return string(b) + "\n"
}
