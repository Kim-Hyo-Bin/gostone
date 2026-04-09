package httperr

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestBadRequest_includesRequestIDWhenSet(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set(GinRequestIDKey, "req-test-123")
	BadRequest(c, "bad thing")
	if w.Code != http.StatusBadRequest {
		t.Fatalf("code %d", w.Code)
	}
	var wrap struct {
		Error struct {
			Code      int    `json:"code"`
			Message   string `json:"message"`
			RequestID string `json:"request_id"`
		} `json:"error"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &wrap); err != nil {
		t.Fatal(err)
	}
	if wrap.Error.RequestID != "req-test-123" {
		t.Fatalf("request_id: %#v", wrap.Error.RequestID)
	}
}

func TestBadRequest_omitsRequestIDWhenUnset(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	BadRequest(c, "x")
	var wrap map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &wrap); err != nil {
		t.Fatal(err)
	}
	errObj, _ := wrap["error"].(map[string]any)
	if errObj == nil {
		t.Fatal("error object")
	}
	if _, ok := errObj["request_id"]; ok {
		t.Fatalf("unexpected request_id: %#v", errObj)
	}
}
