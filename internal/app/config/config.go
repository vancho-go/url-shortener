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
	// LogLevel - уровень логирования для логера.
	LogLevel string
	// EnableHTTPS - включение HTTPS в веб-сервере
	EnableHTTPS bool
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

// WithDSN задает значение для DB DSN.
func (b *serverConfigBuilder) WithDSN(dsn string) *serverConfigBuilder {
	b.config.DBDSN = dsn
	return b
}

// WithLogLevel задает значение для уровня логирования.
func (b *serverConfigBuilder) WithLogLevel(level string) *serverConfigBuilder {
	b.config.LogLevel = level
	return b
}

// WithHTTPS задает значение для https.
func (b *serverConfigBuilder) WithHTTPS(https bool) *serverConfigBuilder {
	b.config.EnableHTTPS = https
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

	var logLevel string
	flag.StringVar(&logLevel, "l", "info", "logger level")

	var enableHTTPS bool
	flag.BoolVar(&enableHTTPS, "s", false, "enable HTTPs on server")

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

	if envLevel := os.Getenv("LOG_LEVEL"); envLevel != "" {
		logLevel = envLevel
	}

	if envEnableHTTPS := os.Getenv("ENABLE_HTTPS"); envEnableHTTPS == "1" {
		enableHTTPS = true
	}

	var builder serverConfigBuilder

	builder.WithServerHost(serverHost).
		WithBaseHost(baseHost).WithFileStorage(fileStorage).WithDSN(dsn).WithLogLevel(logLevel).WithHTTPS(enableHTTPS)

	return builder.config, nil
}
