package buildinfo

import "testing"

func TestVersionNonEmpty(t *testing.T) {
	if Version == "" {
		t.Fatal("Version should default non-empty")
	}
	if Commit == "" {
		t.Fatal("Commit should default non-empty")
	}
}
