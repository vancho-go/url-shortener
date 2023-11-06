package main

import (
	"context"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/vancho-go/url-shortener/internal/app/compress"
	"github.com/vancho-go/url-shortener/internal/app/config"
	"github.com/vancho-go/url-shortener/internal/app/handlers"
	"github.com/vancho-go/url-shortener/internal/app/logger"
	"github.com/vancho-go/url-shortener/internal/app/storage"
	"go.uber.org/zap"
	"net/http"
	"time"
)

const flagLogLevel = "Info"

func initStorage(serverConfig config.ServerConfig) (handlers.Storage, error) {
	switch {
	case serverConfig.DBDSN != "":
		logger.Log.Info("Initializing postgres storage")
		ctx2Sec, cancel2Sec := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel2Sec()
		db, err := storage.Initialize(ctx2Sec, serverConfig.DBDSN)
		if err != nil {
			return nil, errors.New("error Postgres DB initializing")
		}
		return db, nil

	case serverConfig.FileStorage != "":
		logger.Log.Info("Initializing file storage")
		dbInstance, err := storage.NewEncoderDecoder(serverConfig.FileStorage)
		if err != nil {
			return nil, errors.New("error in FileStorage constructor")
		}

		err = dbInstance.Initialize()
		if err != nil {
			return nil, errors.New("error in FileStorage initializing")
		}
		return dbInstance, nil

	default:
		logger.Log.Info("Initializing in-memory storage")
		return storage.MapDB{}, nil
	}
}

func main() {
	err := logger.Initialize(flagLogLevel)
	if err != nil {
		panic(errors.New("error initializing logger"))
	}

	logger.Log.Info("Parsing server configuration")
	configuration, err := config.ParseServer()
	if err != nil {
		panic(errors.New("error parsing server configuration"))
	}

	dbInstance, err := initStorage(configuration)
	if err != nil {
		panic(err)
	}
	defer dbInstance.Close()

	logger.Log.Info("Configuring http compress middleware")
	compressMiddleware := compress.GzipMiddleware

	logger.Log.Info("Running server", zap.String("address", configuration.ServerHost))
	r := chi.NewRouter()
	r.Get("/ping", logger.RequestLogger((handlers.CheckDBConnection(dbInstance))))
	r.Get("/{shortenURL}", logger.RequestLogger(compressMiddleware(handlers.DecodeURL(dbInstance))))
	r.Post("/", logger.RequestLogger(compressMiddleware(handlers.EncodeURL(dbInstance, configuration.BaseHost))))
	r.Post("/api/shorten", logger.RequestLogger(compressMiddleware(handlers.EncodeURLJSON(dbInstance, configuration.BaseHost))))

	err = http.ListenAndServe(configuration.ServerHost, r)
	if err != nil {
		panic(errors.New("error starting server"))
	}
}
