package main

import (
	"github.com/go-chi/chi/v5"
	"github.com/vancho-go/url-shortener/internal/app/handlers"
	"net/http"
)

const addr = "localhost:8080"

func main() {
	r := chi.NewRouter()
	r.Get("/{shortenURL}", handlers.DecodeURL)
	r.Post("/", handlers.EncodeURL(addr))
	err := http.ListenAndServe(addr, r)
	if err != nil {
		panic(err)
	}
}
