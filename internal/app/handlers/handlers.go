package handlers

import (
	"github.com/go-chi/chi/v5"
	"github.com/vancho-go/url-shortener/internal/app/base62"
	"github.com/vancho-go/url-shortener/internal/app/storage"
	"io"
	"math/rand"
	"net/http"
)

var dbInstance = storage.DBInstance

func DecodeURL(res http.ResponseWriter, req *http.Request) {
	shortenURL := chi.URLParam(req, "shortenURL")
	if originalURL, ok := dbInstance[shortenURL]; !ok {
		http.Error(res, "No such shorten URL", http.StatusBadRequest)
		return
	} else {
		res.Header().Set("Location", originalURL)
		res.WriteHeader(http.StatusTemporaryRedirect)
	}
}

func EncodeURL(addr string) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		originalURL, err := io.ReadAll(req.Body)
		if err != nil {
			http.Error(res, err.Error(), http.StatusBadRequest)
			return
		}
		if string(originalURL) == "" {
			res.Header().Set("Content-Type", "text/plain; charset=utf-8")
			http.Error(res, "URL parameter is missing", http.StatusBadRequest)
			return
		}
		id := base62.Base62Encode(rand.Uint64())
		dbInstance[id] = string(originalURL)
		res.WriteHeader(http.StatusCreated)
		shortenURL := addr + "/" + id
		_, _ = res.Write([]byte(shortenURL))
	}
}
