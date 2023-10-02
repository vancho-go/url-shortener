package config

import (
	"flag"
	"os"
)

var Configuration = new(Config)

type Config struct {
	ServerHost string
	BaseHost   string
}

func ParseFlags() {
	flag.StringVar(&Configuration.ServerHost, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&Configuration.BaseHost, "b", "http://"+Configuration.ServerHost, "address and port for shorten URLs")
	flag.Parse()

	if envRunAddr, envBaseAddr := os.Getenv("SERVER_ADDRESS"), os.Getenv("BASE_URL"); envRunAddr != "" && envBaseAddr != "" {
		Configuration.ServerHost = envRunAddr
		Configuration.BaseHost = envBaseAddr
	}
}
