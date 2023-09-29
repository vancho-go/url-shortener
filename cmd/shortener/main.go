package main

import (
	"github.com/vancho-go/url-shortener/internal/app/handlers"
	"net/http"
)

const addr = "localhost:8080"

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc(`/`, handlers.MainPage(addr))
	err := http.ListenAndServe(addr, mux)
	if err != nil {
		panic(err)
	}
}
