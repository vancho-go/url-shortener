package main

import (
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/vancho-go/url-shortener/internal/app/storage"
	"go.uber.org/zap"
	"net/http"

	"github.com/vancho-go/url-shortener/internal/app/auth"
	"github.com/vancho-go/url-shortener/internal/app/compress"
	"github.com/vancho-go/url-shortener/internal/app/config"
	"github.com/vancho-go/url-shortener/internal/app/handlers"
	"github.com/vancho-go/url-shortener/internal/app/logger"
)

//const flagLogLevel = "Info"

func main() {
	configuration, err := config.ParseServer()
	if err != nil {
		panic(errors.New("error parsing server configuration"))
	}

	err = logger.Initialize(configuration.LogLevel)
	if err != nil {
		panic(errors.New("error initializing logger"))
	}

	dbInstance, err := storage.New(configuration)
	if err != nil {
		panic(err)
	}
	defer dbInstance.Close()

	logger.Log.Info("Configuring http compress middleware")
	compressMiddleware := compress.GzipMiddleware

	logger.Log.Info("Running server", zap.String("address", configuration.ServerHost))
	r := chi.NewRouter()

	r.Get("/ping", logger.RequestLogger(handlers.CheckDBConnection(dbInstance)))

	r.Group(func(r chi.Router) {
		r.Use(auth.JWTMiddleware)
		r.Get("/{shortenURL}", logger.RequestLogger(compressMiddleware(handlers.DecodeURL(dbInstance))))
		r.Post("/", logger.RequestLogger(compressMiddleware(handlers.EncodeURL(dbInstance, configuration.BaseHost))))
	})

	r.Route("/api", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(auth.JWTMiddleware)
			r.Post("/shorten", logger.RequestLogger(compressMiddleware(handlers.EncodeURLJSON(dbInstance, configuration.BaseHost))))
			r.Post("/shorten/batch", logger.RequestLogger(compressMiddleware(handlers.EncodeBatch(dbInstance, configuration.BaseHost))))
			r.Get("/user/urls", logger.RequestLogger(handlers.GetUserURLs(dbInstance, configuration.BaseHost)))
			r.Delete("/user/urls", logger.RequestLogger(handlers.DeleteURLs(dbInstance)))
		})

	})

	r.Mount("/debug", handlers.PprofHandler())

	err = http.ListenAndServe(configuration.ServerHost, r)
	if err != nil {
		panic(errors.New("error starting server"))
	}
}
