package bootstrap

import (
	"testing"

	"github.com/Kim-Hyo-Bin/gostone/internal/models"
	"github.com/Kim-Hyo-Bin/gostone/internal/testutil"
)

func TestRunAdmin_refusesNonEmpty(t *testing.T) {
	gdb, err := testutil.OpenMemory()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := testutil.SeedAdmin(gdb, "x"); err != nil {
		t.Fatal(err)
	}
	o := DefaultBootstrapOptions()
	o.AdminPassword = "y"
	if err := RunAdmin(gdb, o); err == nil {
		t.Fatal("expected error")
	}
}

func TestRunAdmin_seeds(t *testing.T) {
	gdb, err := testutil.OpenMemory()
	if err != nil {
		t.Fatal(err)
	}
	o := DefaultBootstrapOptions()
	o.AdminPassword = "secret"
	if err := RunAdmin(gdb, o); err != nil {
		t.Fatal(err)
	}
	var n int64
	if err := gdb.Model(&models.User{}).Count(&n).Error; err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("users %d", n)
	}
}
