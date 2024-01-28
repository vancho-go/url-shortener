package main

import (
	"errors"
	"net/http"
	"net/http/pprof"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/vancho-go/url-shortener/internal/app/auth"
	"github.com/vancho-go/url-shortener/internal/app/compress"
	"github.com/vancho-go/url-shortener/internal/app/config"
	"github.com/vancho-go/url-shortener/internal/app/handlers"
	"github.com/vancho-go/url-shortener/internal/app/logger"
	"github.com/vancho-go/url-shortener/internal/app/storage"
)

const flagLogLevel = "Info"

func initStorage(serverConfig config.ServerConfig) (handlers.Storager, error) {
	switch {
	case serverConfig.DBDSN != "":
		logger.Log.Info("Initializing postgres storage")
		db, err := storage.Initialize(serverConfig.DBDSN)
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

	r.Mount("/debug", pprofHandler())

	err = http.ListenAndServe(configuration.ServerHost, r)
	if err != nil {
		panic(errors.New("error starting server"))
	}
}

func pprofHandler() http.Handler {
	r := chi.NewRouter()
	r.Get("/", pprof.Index)
	r.Get("/cmdline", pprof.Cmdline)
	r.Get("/profile", pprof.Profile)
	r.Get("/symbol", pprof.Symbol)
	r.Get("/trace", pprof.Trace)
	r.Get("/allocs", pprof.Handler("allocs").ServeHTTP)
	r.Get("/block", pprof.Handler("block").ServeHTTP)
	r.Get("/goroutine", pprof.Handler("goroutine").ServeHTTP)
	r.Get("/heap", pprof.Handler("heap").ServeHTTP)
	r.Get("/mutex", pprof.Handler("mutex").ServeHTTP)
	r.Get("/threadcreate", pprof.Handler("threadcreate").ServeHTTP)

	return r
}
