package main

import (
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/vancho-go/url-shortener/internal/app/config"
	"github.com/vancho-go/url-shortener/internal/app/handlers"
	"github.com/vancho-go/url-shortener/internal/app/storage"
	"net/http"
)

var dbInstance = make(storage.MapDB)

func main() {
	config.ParseServerFlags()

	r := chi.NewRouter()
	r.Get("/{shortenURL}", handlers.DecodeURL(dbInstance))
	r.Post("/", handlers.EncodeURL(dbInstance, config.Configuration.BaseHost))

	err := http.ListenAndServe(config.Configuration.ServerHost, r)
	if err != nil {
		panic(errors.New("error starting server"))
	}
}
