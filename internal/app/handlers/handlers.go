package handlers

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"github.com/vancho-go/url-shortener/internal/app/base62"
	"github.com/vancho-go/url-shortener/internal/app/logger"
	"github.com/vancho-go/url-shortener/internal/app/models"
	"go.uber.org/zap"
	"io"
	"math/rand"
	"net/http"
)

type Storage interface {
	AddURL(string, string) error
	GetURL(string) (string, error)
}

func DecodeURL(db Storage) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		shortenURL := chi.URLParam(req, "shortenURL")
		originalURL, err := db.GetURL(shortenURL)
		if err != nil {
			http.Error(res, "No such shorten URL", http.StatusBadRequest)
			return
		}
		res.Header().Set("Location", originalURL)
		res.WriteHeader(http.StatusTemporaryRedirect)

	}
}

func EncodeURL(db Storage, addr string) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		originalURL, err := io.ReadAll(req.Body)
		if err != nil {
			http.Error(res, "No such shorten URL", http.StatusBadRequest)
			return
		}
		if string(originalURL) == "" {
			//res.Header().Set("Content-Type", "text/plain; charset=utf-8")
			http.Error(res, "URL parameter is missing", http.StatusBadRequest)
			return
		}

		shortenURL := base62.Base62Encode(rand.Uint64())
		err = db.AddURL(string(originalURL), shortenURL)
		if err != nil {
			http.Error(res, "Error adding new shorten URL", http.StatusBadRequest)
			return
		}
		res.WriteHeader(http.StatusCreated)
		_, err = res.Write([]byte(addr + "/" + shortenURL))
		if err != nil {
			logger.Log.Error("write failed", zap.Error(err))
		}
	}
}

func EncodeURLJSON(db Storage, addr string) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		var request models.APIShortenRequest
		dec := json.NewDecoder(req.Body)
		if err := dec.Decode(&request); err != nil {
			logger.Log.Debug("can't decode request JSON body", zap.Error(err))
			http.Error(res, "Error adding new shorten URL", http.StatusBadRequest)
			return
		}

		shortenURL := base62.Base62Encode(rand.Uint64())
		originalURL := request.URL
		if originalURL == "" {
			http.Error(res, "URL parameter is missing", http.StatusBadRequest)
			return
		}
		err := db.AddURL(string(originalURL), shortenURL)
		if err != nil {
			http.Error(res, "Error adding new shorten URL", http.StatusBadRequest)
			return
		}

		response := models.APIShortenResponse{
			Result: addr + "/" + shortenURL,
		}

		res.Header().Set("Content-Type", "application/json")
		res.WriteHeader(http.StatusCreated)
		enc := json.NewEncoder(res)
		if err := enc.Encode(response); err != nil {
			logger.Log.Debug("error encoding response", zap.Error(err))
			http.Error(res, "Error adding new shorten URL", http.StatusBadRequest)
			return
		}
	}
}
