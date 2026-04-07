package app

import (
	"strings"
	"testing"

	"github.com/Kim-Hyo-Bin/gostone/internal/conf"
)

func TestRun_rejectsEmptyTokenSecret(t *testing.T) {
	cfg := &conf.Config{}
	cfg.Database.Connection = "file::memory:?cache=shared"
	cfg.Service.Listen = "127.0.0.1:0"
	err := Run(cfg)
	if err == nil || !strings.Contains(err.Error(), "secret") {
		t.Fatalf("got %v", err)
	}
}
