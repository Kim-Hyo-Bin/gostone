package main

import (
	"log"

	"github.com/Kim-Hyo-Bin/gostone/internal/app"
	"github.com/Kim-Hyo-Bin/gostone/internal/config"
)

func main() {
	opts := config.ParseFlags()
	cfg, err := config.Load(opts)
	if err != nil {
		log.Fatal(err)
	}
	if err := app.Run(cfg); err != nil {
		log.Fatal(err)
	}
}
