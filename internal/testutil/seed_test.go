package testutil

import "testing"

func TestOpenMemory(t *testing.T) {
	gdb, err := OpenMemory()
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := gdb.DB()
	if err != nil {
		t.Fatal(err)
	}
	if err := sqlDB.Ping(); err != nil {
		t.Fatal(err)
	}
}
