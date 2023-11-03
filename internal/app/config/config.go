package config

import (
	"flag"
	"os"
)

type ServerConfig struct {
	ServerHost  string
	BaseHost    string
	FileStorage string
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

func (b *serverConfigBuilder) WithFileStorage(fileStorage string) *serverConfigBuilder {
	b.config.FileStorage = fileStorage
	return b
}

func ParseServer() (ServerConfig, error) {
	var serverHost string
	flag.StringVar(&serverHost, "a", "localhost:8080", "address and port to run server")

	var baseHost string
	flag.StringVar(&baseHost, "b", "http://localhost:8080", "address and port for shorten URLs")

	var fileStorage string
	flag.StringVar(&fileStorage, "f", "/tmp/short-url-db.json", "absolute path for file storage")

	flag.Parse()

	if envRunAddr := os.Getenv("SERVER_ADDRESS"); envRunAddr != "" {
		serverHost = envRunAddr
	}

	if envBaseAddr := os.Getenv("BASE_URL"); envBaseAddr != "" {
		baseHost = envBaseAddr
	}

	if envFileStorage := os.Getenv("FILE_STORAGE_PATH"); envFileStorage != "" {
		fileStorage = envFileStorage
	}

	var builder serverConfigBuilder

	builder.WithServerHost(serverHost).
		WithBaseHost(baseHost).WithFileStorage(fileStorage)

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

type DBConfig struct {
	DSN string
}

type dbConfigBuilder struct {
	config DBConfig
}

func (b *dbConfigBuilder) WithDSN(dsn string) *dbConfigBuilder {
	b.config.DSN = dsn
	return b
}

func ParseDB() (DBConfig, error) {
	var dsn string
	flag.StringVar(&dsn, "d", "host=localhost port=5432 user=vancho password=vancho_pswd dbname=vancho_db sslmode=disable", "data source name for driver to connect to DB")

	flag.Parse()

	if envDSN := os.Getenv("DATABASE_DSN"); envDSN != "" {
		dsn = envDSN
	}

	var builder dbConfigBuilder

	builder.WithDSN(dsn)

	return builder.config, nil
}
