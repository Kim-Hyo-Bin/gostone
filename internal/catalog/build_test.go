package catalog

import (
	"testing"

	"github.com/Kim-Hyo-Bin/gostone/internal/db"
)

func TestBuild_empty(t *testing.T) {
	gdb, err := db.Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(gdb); err != nil {
		t.Fatal(err)
	}
	out, err := Build(gdb)
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 0 {
		t.Fatalf("got %d", len(out))
	}
}
