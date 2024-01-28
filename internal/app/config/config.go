// Модуль config собирает информацию о параметрах, необходимых для инициализации сервера и клиента.
package config

import (
	"flag"
	"os"
)

// ServerConfig хранит параметры, необходимые для инициализации сервера.
type ServerConfig struct {
	// ServerHost - адрес и порт, на котором будет поднят сервер.
	ServerHost string
	// BaseHost - адрес и порт для shortenURLs.
	BaseHost string
	// FileStorage - путь к файлу json, в котором будут храниться данные, если в качестве хранилища данных используется файл.
	FileStorage string
	// DBDSN - connection string для подключения к БД.
	DBDSN string
}

type serverConfigBuilder struct {
	config ServerConfig
}

// WithServerHost задает значение для ServerHost.
func (b *serverConfigBuilder) WithServerHost(serverHost string) *serverConfigBuilder {
	b.config.ServerHost = serverHost
	return b
}

// WithBaseHost задает значение для BaseHost.
func (b *serverConfigBuilder) WithBaseHost(baseHost string) *serverConfigBuilder {
	b.config.BaseHost = baseHost
	return b
}

// WithFileStorage задает значение для FileStorage.
func (b *serverConfigBuilder) WithFileStorage(fileStorage string) *serverConfigBuilder {
	b.config.FileStorage = fileStorage
	return b
}

// WithDSN задает значение для DBDSN.
func (b *serverConfigBuilder) WithDSN(dsn string) *serverConfigBuilder {
	b.config.DBDSN = dsn
	return b
}

// ParseServer генерирует конфигурацию для инициализации сервера.
func ParseServer() (ServerConfig, error) {
	var serverHost string
	flag.StringVar(&serverHost, "a", "localhost:8080", "address and port to run server")

	var baseHost string
	flag.StringVar(&baseHost, "b", "http://localhost:8080", "address and port for shorten URLs")

	var fileStorage string
	flag.StringVar(&fileStorage, "f", "/tmp/short-url-db.json", "absolute path for file storage")

	var dsn string
	flag.StringVar(&dsn, "d", "", "data source name for driver to connect to DB")

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

	if envDSN := os.Getenv("DATABASE_DSN"); envDSN != "" {
		dsn = envDSN
	}

	var builder serverConfigBuilder

	builder.WithServerHost(serverHost).
		WithBaseHost(baseHost).WithFileStorage(fileStorage).WithDSN(dsn)

	return builder.config, nil
}

// ClientConfig хранит параметры, необходимые для инициализации клиента.
type ClientConfig struct {
	// ClientHost - адрес и порт для клиента.
	ClientHost string
}

type clientConfigBuilder struct {
	config ClientConfig
}

// WithClientHost задает значение для ClientHost.
func (b *clientConfigBuilder) WithClientHost(clientHost string) *clientConfigBuilder {
	b.config.ClientHost = clientHost
	return b
}

// ParseClient генерирует конфигурацию для инициализации клиента.
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
