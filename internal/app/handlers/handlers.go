package handlers

import (
	"github.com/vancho-go/url-shortener/internal/app/base62"
	"github.com/vancho-go/url-shortener/internal/app/storage"
	"io"
	"math/rand"
	"net/http"
	url2 "net/url"
	"strings"
)

var dbInstance = storage.DBInstance

func MainPage(addr string) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		switch req.Method {
		case http.MethodGet:
			decodeURL(res, req)
		case http.MethodPost:
			encodeURL("localhost:8080", res, req)
		}
	}
}

func decodeURL(res http.ResponseWriter, req *http.Request) {
	shortenURL := strings.TrimPrefix(req.URL.Path, "/")
	if originalURL, ok := dbInstance[shortenURL]; !ok {
		http.Error(res, "No such shorten URL", http.StatusBadRequest)
		return
	} else {
		res.Header().Set("Location", originalURL)
		res.WriteHeader(http.StatusTemporaryRedirect)
	}
}

func encodeURL(addr string, res http.ResponseWriter, req *http.Request) {
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
	shortenURL, err := url2.JoinPath("http://", addr, id)
	if err != nil {
		res.Header().Set("Content-Type", "text/plain; charset=utf-8")
		http.Error(res, err.Error(), http.StatusBadRequest)
		return
	}
	_, _ = res.Write([]byte(shortenURL))
}
