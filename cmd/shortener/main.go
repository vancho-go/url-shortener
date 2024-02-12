package main

import (
	"fmt"
	"github.com/vancho-go/url-shortener/internal/app"
	"log"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func main() {
	printBuildInfo()

	if err := app.Run(); err != nil {
		log.Panic(err.Error())
	}
}

func printBuildInfo() {
	fmt.Println("Build version: %s", buildVersion)
	fmt.Println("Build date: %s", buildDate)
	fmt.Println("Build commit: %s", buildCommit)
}
