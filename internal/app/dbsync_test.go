package app

import (
	"fmt"
	"testing"

	"github.com/Kim-Hyo-Bin/gostone/internal/conf"
	"github.com/google/uuid"
)

func TestDBSync_memory(t *testing.T) {
	dsn := fmt.Sprintf("file:dbsync_%s?mode=memory&cache=shared", uuid.NewString())
	cfg := &conf.Config{}
	cfg.Database.Connection = dsn
	if err := DBSync(cfg); err != nil {
		t.Fatal(err)
	}
	if err := DBSync(cfg); err != nil {
		t.Fatal("second sync should be idempotent", err)
	}
}
