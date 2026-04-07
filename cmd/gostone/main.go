package main

import (
	"log"

	"github.com/Kim-Hyo-Bin/gostone/internal/app"
)

func main() {
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
