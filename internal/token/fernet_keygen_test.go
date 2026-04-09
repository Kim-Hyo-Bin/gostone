package token

import (
	"testing"
)

func TestWriteNextKeystoneFernetKey_emptyDir(t *testing.T) {
	dir := t.TempDir()
	id, err := WriteNextKeystoneFernetKey(dir)
	if err != nil {
		t.Fatal(err)
	}
	if id != 0 {
		t.Fatalf("first key id: %d", id)
	}
	keys, err := LoadFernetKeysFromRepo(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(keys) != 1 {
		t.Fatalf("keys %d", len(keys))
	}
}

func TestWriteNextKeystoneFernetKey_increments(t *testing.T) {
	dir := t.TempDir()
	if _, err := WriteNextKeystoneFernetKey(dir); err != nil {
		t.Fatal(err)
	}
	id, err := WriteNextKeystoneFernetKey(dir)
	if err != nil {
		t.Fatal(err)
	}
	if id != 1 {
		t.Fatalf("next id: %d", id)
	}
	keys, err := LoadFernetKeysFromRepo(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(keys) != 2 {
		t.Fatalf("want 2 keys, got %d", len(keys))
	}
}
