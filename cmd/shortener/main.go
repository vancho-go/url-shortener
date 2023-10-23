package main

import (
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/vancho-go/url-shortener/internal/app/compress"
	"github.com/vancho-go/url-shortener/internal/app/config"
	"github.com/vancho-go/url-shortener/internal/app/handlers"
	"github.com/vancho-go/url-shortener/internal/app/logger"
	"github.com/vancho-go/url-shortener/internal/app/storage"
	"go.uber.org/zap"
	"net/http"
)

var dbInstance = make(storage.MapDB)

const flagLogLevel = "Info"

func main() {
	err := logger.Initialize(flagLogLevel)
	if err != nil {
		panic(errors.New("error initializing logger"))
	}

	logger.Log.Info("Parsing server config")
	config, err := config.ParseServer()
	if err != nil {
		panic(errors.New("error parsing server config"))
	}

	logger.Log.Info("Configuring http compress middleware")
	compressMiddleware := compress.GzipMiddleware

	logger.Log.Info("Running server", zap.String("address", config.ServerHost))
	r := chi.NewRouter()
	r.Get("/{shortenURL}", logger.RequestLogger(compressMiddleware(handlers.DecodeURL(dbInstance))))
	r.Post("/", logger.RequestLogger(compressMiddleware(handlers.EncodeURL(dbInstance, config.BaseHost))))
	r.Post("/api/shorten", logger.RequestLogger(compressMiddleware(handlers.EncodeURLJSON(dbInstance, config.BaseHost))))

	err = http.ListenAndServe(config.ServerHost, r)
	if err != nil {
		panic(errors.New("error starting server"))
	}
}
