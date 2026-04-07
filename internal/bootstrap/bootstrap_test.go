package bootstrap

import (
	"testing"

	"github.com/Kim-Hyo-Bin/gostone/internal/models"
	"github.com/Kim-Hyo-Bin/gostone/internal/testutil"
)

func TestFromEnv_noPassword_skips(t *testing.T) {
	t.Setenv("GOSTONE_BOOTSTRAP_ADMIN_PASSWORD", "")
	gdb, err := testutil.OpenMemory()
	if err != nil {
		t.Fatal(err)
	}
	if err := FromEnv(gdb); err != nil {
		t.Fatal(err)
	}
	var n int64
	if err := gdb.Model(&models.User{}).Count(&n).Error; err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatalf("expected no users, got %d", n)
	}
}

func TestFromEnv_seedsOnce(t *testing.T) {
	t.Setenv("GOSTONE_BOOTSTRAP_ADMIN_PASSWORD", "boot-pw")
	gdb, err := testutil.OpenMemory()
	if err != nil {
		t.Fatal(err)
	}
	if err := FromEnv(gdb); err != nil {
		t.Fatal(err)
	}
	var n int64
	if err := gdb.Model(&models.User{}).Count(&n).Error; err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("users: %d", n)
	}
	if err := FromEnv(gdb); err != nil {
		t.Fatal(err)
	}
	if err := gdb.Model(&models.User{}).Count(&n).Error; err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("second call should not duplicate: %d", n)
	}
}
