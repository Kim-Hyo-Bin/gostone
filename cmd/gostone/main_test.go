package main

import "testing"

// Smoke test so cmd/gostone participates in coverage merges and `go test ./...` is uniform.
func TestMainPackageBuilds(t *testing.T) {
	t.Parallel()
}
