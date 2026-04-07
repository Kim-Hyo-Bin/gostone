package auth

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestFromGin_missing(t *testing.T) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	_, ok := FromGin(c)
	if ok {
		t.Fatal()
	}
}

func TestFromGin_present(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	want := Context{UserID: "u", DomainID: "d", ProjectID: "p", Roles: []string{"admin"}}
	c.Set(GinKey, want)
	got, ok := FromGin(c)
	if !ok || got.UserID != want.UserID || !got.HasRole("admin") || got.HasRole("nope") {
		t.Fatalf("got %+v ok=%v", got, ok)
	}
}

func TestFromGin_wrongType(t *testing.T) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Set(GinKey, "string")
	_, ok := FromGin(c)
	if ok {
		t.Fatal()
	}
}

func TestContext_HasRole(t *testing.T) {
	c := Context{Roles: []string{"a", "b"}}
	if !c.HasRole("a") || c.HasRole("z") {
		t.Fatal()
	}
}
