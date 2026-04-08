package app

import (
	"fmt"
	"testing"

	"github.com/Kim-Hyo-Bin/gostone/internal/bootstrap"
	"github.com/Kim-Hyo-Bin/gostone/internal/conf"
	"github.com/Kim-Hyo-Bin/gostone/internal/db"
	"github.com/Kim-Hyo-Bin/gostone/internal/models"
	"github.com/google/uuid"
)

func TestBootstrap_memory(t *testing.T) {
	dsn := fmt.Sprintf("file:boot_%s?mode=memory&cache=shared", uuid.NewString())
	cfg := &conf.Config{}
	cfg.Database.Connection = dsn
	o := bootstrap.DefaultBootstrapOptions()
	o.AdminPassword = "bootstrap-test-secret"
	if err := Bootstrap(cfg, o); err != nil {
		t.Fatal(err)
	}
	if err := Bootstrap(cfg, o); err == nil {
		t.Fatal("second bootstrap should refuse non-empty DB")
	}
}

func TestBootstrap_catalogRegionFromConfig(t *testing.T) {
	dsn := fmt.Sprintf("file:bootreg_%s?mode=memory&cache=shared", uuid.NewString())
	cfg := &conf.Config{}
	cfg.Database.Connection = dsn
	cfg.Service.RegionID = "FromConfig"
	o := bootstrap.DefaultBootstrapOptions()
	o.AdminPassword = "x"
	if err := Bootstrap(cfg, o); err != nil {
		t.Fatal(err)
	}
	gdb, err := db.Open(dsn)
	if err != nil {
		t.Fatal(err)
	}
	var reg models.Region
	if err := gdb.First(&reg).Error; err != nil {
		t.Fatal(err)
	}
	if reg.ID != "FromConfig" {
		t.Fatalf("region id: %q", reg.ID)
	}
}
