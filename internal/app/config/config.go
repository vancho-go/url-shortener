package config

import (
	"flag"
	"os"
)

type ServerConfig struct {
	ServerHost string
	BaseHost   string
}

type serverConfigBuilder struct {
	config ServerConfig
}

func (b *serverConfigBuilder) WithServerHost(serverHost string) *serverConfigBuilder {
	b.config.ServerHost = serverHost
	return b
}

func (b *serverConfigBuilder) WithBaseHost(baseHost string) *serverConfigBuilder {
	b.config.BaseHost = baseHost
	return b
}

func ParseServer() (ServerConfig, error) {
	var serverHost string
	flag.StringVar(&serverHost, "a", "localhost:8080", "address and port to run server")

	var baseHost string
	flag.StringVar(&baseHost, "b", "http://localhost:8080", "address and port for shorten URLs")

	flag.Parse()

	if envRunAddr, envBaseAddr := os.Getenv("SERVER_ADDRESS"), os.Getenv("BASE_URL"); envRunAddr != "" && envBaseAddr != "" {
		serverHost = envRunAddr
		baseHost = envBaseAddr
	}

	var builder serverConfigBuilder

	builder.WithServerHost(serverHost).
		WithBaseHost(baseHost)

	return builder.config, nil
}

type ClientConfig struct {
	ClientHost string
}

type clientConfigBuilder struct {
	config ClientConfig
}

func (b *clientConfigBuilder) WithClientHost(clientHost string) *clientConfigBuilder {
	b.config.ClientHost = clientHost
	return b
}
func ParseClient() (ClientConfig, error) {
	var clientHost string
	flag.StringVar(&clientHost, "c", "http://localhost:8080", "address and port for client")

	flag.Parse()

	if envClientAddr := os.Getenv("CLIENT_ADDRESS"); envClientAddr != "" {
		clientHost = envClientAddr
	}

	var builder clientConfigBuilder

	builder.WithClientHost(clientHost)

	return builder.config, nil
}
