package main

import (
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/vancho-go/url-shortener/internal/app/config"
	"github.com/vancho-go/url-shortener/internal/app/handlers"
	"net/http"
)

func main() {
	config.ParseServerFlags()

	r := chi.NewRouter()
	r.Get("/{shortenURL}", handlers.DecodeURL)
	r.Post("/", handlers.EncodeURL(config.Configuration.BaseHost))

	err := http.ListenAndServe(config.Configuration.ServerHost, r)
	if err != nil {
		panic(errors.New("error starting server"))
	}
}
