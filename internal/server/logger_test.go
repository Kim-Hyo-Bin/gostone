package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestJSONAccessLogFormatter_includesHostAndUserAgent(t *testing.T) {
	gin.SetMode(gin.TestMode)
	req := httptest.NewRequest(http.MethodGet, "/v3/users", nil)
	req.Host = "keystone.example:5000"
	req.Header.Set("User-Agent", "python-keystoneclient/5.0")

	line := jsonAccessLogFormatter(gin.LogFormatterParams{
		Request:    req,
		TimeStamp:  time.Unix(0, 0).UTC(),
		StatusCode: 200,
		Latency:    time.Millisecond,
		ClientIP:   "192.0.2.1",
		Method:     "GET",
		Path:       "/v3/users",
		BodySize:   42,
		Keys:       map[any]any{requestIDGinKey: "req-xyz"},
	})

	var m map[string]any
	if err := json.Unmarshal([]byte(line[:len(line)-1]), &m); err != nil {
		t.Fatal(err)
	}
	if m["host"] != "keystone.example:5000" {
		t.Fatalf("host: %v", m["host"])
	}
	if m["user_agent"] != "python-keystoneclient/5.0" {
		t.Fatalf("user_agent: %v", m["user_agent"])
	}
	if m["request_id"] != "req-xyz" {
		t.Fatalf("request_id: %v", m["request_id"])
	}
}
