// Модуль config собирает информацию о параметрах, необходимых для инициализации сервера и клиента.
package config

import (
	"encoding/json"
	"flag"
	"os"
)

// JSONConfig - cтруктура, соответствующая JSON файлу.
type JSONConfig struct {
	ServerAddress   string `json:"server_address"`
	BaseURL         string `json:"base_url"`
	FileStoragePath string `json:"file_storage_path"`
	DatabaseDSN     string `json:"database_dsn"`
	EnableHTTPS     bool   `json:"enable_https"`
}

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
func ParseServer() (*ServerConfig, error) {
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

	var jsonConfigFile string
	flag.StringVar(&jsonConfigFile, "c", "", "absolute path for json config file")
	flag.StringVar(&jsonConfigFile, "config", "", "path for json config file")

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

	if envJSONConfigFile := os.Getenv("CONFIG"); envJSONConfigFile != "" {
		jsonConfigFile = envJSONConfigFile
	}

	if jsonConfigFile != "" {
		jsonConfig, err := parseJSONConfig(jsonConfigFile)
		if err != nil {
			return nil, err
		}

		if serverHost == "" {
			serverHost = jsonConfig.ServerAddress
		}
		if baseHost == "" {
			baseHost = jsonConfig.BaseURL
		}
		if fileStorage == "" {
			fileStorage = jsonConfig.FileStoragePath
		}
		if dsn == "" {
			dsn = jsonConfig.DatabaseDSN
		}
		if !enableHTTPS && jsonConfig.EnableHTTPS {
			enableHTTPS = jsonConfig.EnableHTTPS
		}
	}

	var builder serverConfigBuilder

	builder.WithServerHost(serverHost).
		WithBaseHost(baseHost).
		WithFileStorage(fileStorage).
		WithDSN(dsn).
		WithLogLevel(logLevel).
		WithHTTPS(enableHTTPS)

	return &builder.config, nil
}

// parseJSONConfig считывает конфигурацию из json файла
func parseJSONConfig(configFile string) (*JSONConfig, error) {
	file, err := os.Open(configFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var config JSONConfig
	decoder := json.NewDecoder(file)
	if err = decoder.Decode(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
