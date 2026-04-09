package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestSubjectOrBearerToken_prefersSubject(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodGet, "/x", nil)
	c.Request.Header.Set("Authorization", "Bearer caller-token")
	c.Request.Header.Set("X-Subject-Token", "subject-token")
	if got := SubjectOrBearerToken(c); got != "subject-token" {
		t.Fatalf("got %q", got)
	}
}
