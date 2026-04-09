package db

import (
	"testing"
)

func TestOpenAndMigrate(t *testing.T) {
	gdb, err := Open("file::memory:?cache=shared")
	if err != nil {
		t.Fatal(err)
	}
	if err := AutoMigrate(gdb); err != nil {
		t.Fatal(err)
	}
}
