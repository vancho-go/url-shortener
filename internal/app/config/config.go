package config

import "flag"

var Configuration = new(Config)

type Config struct {
	ServerHost string
	BaseHost   string
}

func ParseFlags() {
	flag.StringVar(&Configuration.ServerHost, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&Configuration.BaseHost, "b", Configuration.ServerHost, "address and port for shorten URLs")
	flag.Parse()
}
