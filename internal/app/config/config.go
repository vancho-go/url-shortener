package config

import (
	"flag"
	"os"
)

var Configuration = new(Config)

type Config struct {
	ServerHost string
	ClientHost string
	BaseHost   string
}

func ParseServerFlags() {
	flag.StringVar(&Configuration.ServerHost, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&Configuration.BaseHost, "b", "http://localhost:8080", "address and port for shorten URLs")
	flag.Parse()

	if envRunAddr, envBaseAddr, envClientAddr := os.Getenv("SERVER_ADDRESS"), os.Getenv("BASE_URL"), os.Getenv("CLIENT_ADDRESS"); envRunAddr != "" && envBaseAddr != "" && envClientAddr != "" {
		Configuration.ServerHost = envRunAddr
		Configuration.BaseHost = envBaseAddr
		Configuration.ClientHost = envClientAddr
	}
}

func ParseClientFlags() {
	flag.StringVar(&Configuration.ClientHost, "c", "http://localhost:8080", "address and port for client")
	flag.Parse()

	if envClientAddr := os.Getenv("CLIENT_ADDRESS"); envClientAddr != "" {
		Configuration.ClientHost = envClientAddr
	}
}
