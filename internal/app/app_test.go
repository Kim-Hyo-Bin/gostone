package app

import (
	"strings"
	"testing"

	"github.com/Kim-Hyo-Bin/gostone/internal/conf"
)

func TestRun_rejectsEmptyTokenSecretWhenJWT(t *testing.T) {
	cfg := &conf.Config{}
	cfg.Database.Connection = "file::memory:?cache=shared"
	cfg.Service.Listen = "127.0.0.1:0"
	cfg.Token.Provider = "jwt"
	err := Run(cfg)
	if err == nil || !strings.Contains(err.Error(), "secret") {
		t.Fatalf("got %v", err)
	}
}
