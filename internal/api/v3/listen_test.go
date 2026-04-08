package v3

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Kim-Hyo-Bin/gostone/internal/listenctx"
	"github.com/gin-gonic/gin"
)

func TestListenBinding_fromContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	c.Request = req.WithContext(listenctx.WithLabel(req.Context(), "admin"))
	if ListenBinding(c) != "admin" {
		t.Fatal()
	}
}

func TestListenBinding_empty(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	if ListenBinding(c) != "" {
		t.Fatal()
	}
}
