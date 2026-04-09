// Package listenctx stores which HTTP bind (public, admin, internal, listen) served a request.
package listenctx

import (
	"context"
	"net/http"
)

type labelKey struct{}

// WithLabel returns ctx carrying the listen binding name (e.g. public, admin).
func WithLabel(ctx context.Context, name string) context.Context {
	return context.WithValue(ctx, labelKey{}, name)
}

// Label returns the binding name, or empty if unset (single-stack callers may omit).
func Label(ctx context.Context) string {
	s, _ := ctx.Value(labelKey{}).(string)
	return s
}

// WrapHandler sets the label on the request context before delegating to h.
func WrapHandler(label string, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r.WithContext(WithLabel(r.Context(), label)))
	})
}
