package main

import (
	"log"

	"github.com/Kim-Hyo-Bin/gostone/internal/app"
	"github.com/Kim-Hyo-Bin/gostone/internal/conf"
)

func main() {
	opts := conf.ParseFlags()
	cfg, err := conf.Load(opts)
	if err != nil {
		log.Fatal(err)
	}
	if err := app.Run(cfg); err != nil {
		log.Fatal(err)
	}
}
