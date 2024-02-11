package main

import (
	"github.com/vancho-go/url-shortener/internal/app"
	"log"
)

func main() {
	if err := app.Run(); err != nil {
		log.Panic(err.Error())
	}
}
