package listenctx

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWrapHandler_setsLabel(t *testing.T) {
	var got string
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = Label(r.Context())
	})
	srv := httptest.NewServer(WrapHandler("public", h))
	defer srv.Close()
	resp, err := http.Get(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()
	if got != "public" {
		t.Fatalf("label %q", got)
	}
}

func TestLabel_empty(t *testing.T) {
	if Label(context.Background()) != "" {
		t.Fatal()
	}
}
