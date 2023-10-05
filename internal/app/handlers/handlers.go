package handlers

import (
	"github.com/go-chi/chi/v5"
	"github.com/vancho-go/url-shortener/internal/app/base62"
	"github.com/vancho-go/url-shortener/internal/app/storage"
	"io"
	"log"
	"math/rand"
	"net/http"
)

var dbInstance = make(storage.DBInstance)

func DecodeURL(res http.ResponseWriter, req *http.Request) {
	shortenURL := chi.URLParam(req, "shortenURL")
	originalURL, err := dbInstance.GetURL(shortenURL)
	if err != nil {
		http.Error(res, "No such shorten URL", http.StatusBadRequest)
		return
	}
	res.Header().Set("Location", originalURL)
	res.WriteHeader(http.StatusTemporaryRedirect)

}

func EncodeURL(addr string) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		originalURL, err := io.ReadAll(req.Body)
		if err != nil {
			http.Error(res, "Error reading body", http.StatusBadRequest)
			return
		}
		if string(originalURL) == "" {
			res.Header().Set("Content-Type", "text/plain; charset=utf-8")
			http.Error(res, "URL parameter is missing", http.StatusBadRequest)
			return
		}

		shortenURL := base62.Base62Encode(rand.Uint64())
		err = dbInstance.AddURL(string(originalURL), shortenURL)
		if err != nil {
			http.Error(res, "Error adding new shorten URL", http.StatusBadRequest)
			return
		}
		res.WriteHeader(http.StatusCreated)
		_, err = res.Write([]byte(addr + "/" + shortenURL))
		if err != nil {
			log.Printf("Write failed: %v", err)
		}
	}
}
