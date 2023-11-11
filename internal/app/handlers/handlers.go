package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/vancho-go/url-shortener/internal/app/base62"
	"github.com/vancho-go/url-shortener/internal/app/logger"
	"github.com/vancho-go/url-shortener/internal/app/models"
	"github.com/vancho-go/url-shortener/internal/app/storage"
	"go.uber.org/zap"
	"io"
	"math/rand"
	"net/http"
	"time"
)

var ErrUnique = errors.New("original URL already exists")

type Storage interface {
	AddURL(context.Context, string, string) error
	GetURL(context.Context, string) (string, error)
	IsShortenUnique(context.Context, string) bool
	Close() error
}

func DecodeURL(db Storage) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		shortenURL := chi.URLParam(req, "shortenURL")
		ctx, cancel := context.WithTimeout(req.Context(), 3*time.Second)
		defer cancel()
		originalURL, err := db.GetURL(ctx, shortenURL)
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
			http.Error(res, "URL parameter is missing", http.StatusBadRequest)
			return
		}

		shortenURL := base62.Base62Encode(rand.Uint64())
		ctx, cancel := context.WithTimeout(req.Context(), 3*time.Second)
		defer cancel()
		for !db.IsShortenUnique(ctx, shortenURL) {
			shortenURL = base62.Base62Encode(rand.Uint64())
		}

		ctx, cancel2 := context.WithTimeout(req.Context(), 3*time.Second)
		defer cancel2()
		err = db.AddURL(ctx, string(originalURL), shortenURL)
		if err != nil {
			if !isUniqueViolationError(err) {
				http.Error(res, "Error adding new shorten URL", http.StatusBadRequest)
				return
			}

			pg, ok := db.(*storage.Database)
			if !ok {
				http.Error(res, "Internal DB Error", http.StatusInternalServerError)
				return
			}
			ctx, cancel3 := context.WithTimeout(req.Context(), 3*time.Second)
			defer cancel3()
			shortenURL, err = pg.GetShortenURLByOriginal(ctx, string(originalURL))
			if err != nil {
				http.Error(res, "Error getting shorten URL", http.StatusInternalServerError)
				return
			}
			res.WriteHeader(http.StatusConflict)
		} else {
			res.WriteHeader(http.StatusCreated)
		}

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

		originalURL := request.URL
		if originalURL == "" {
			http.Error(res, "URL parameter is missing", http.StatusBadRequest)
			return
		}

		shortenURL := base62.Base62Encode(rand.Uint64())
		ctx, cancel := context.WithTimeout(req.Context(), 3*time.Second)
		defer cancel()
		for !db.IsShortenUnique(ctx, shortenURL) {
			shortenURL = base62.Base62Encode(rand.Uint64())
		}

		ctx, cancel2 := context.WithTimeout(req.Context(), 3*time.Second)
		defer cancel2()
		err := db.AddURL(ctx, originalURL, shortenURL)
		if err != nil {
			if !isUniqueViolationError(err) {
				http.Error(res, "Error adding new shorten URL", http.StatusBadRequest)
				return
			}

			pg, ok := db.(*storage.Database)
			if !ok {
				http.Error(res, "Internal DB Error", http.StatusInternalServerError)
				return
			}
			ctx, cancel3 := context.WithTimeout(req.Context(), 3*time.Second)
			defer cancel3()
			shortenURL, err = pg.GetShortenURLByOriginal(ctx, originalURL)
			if err != nil {
				http.Error(res, "Error getting shorten URL", http.StatusInternalServerError)
				return
			}
			res.Header().Set("Content-Type", "application/json")
			res.WriteHeader(http.StatusConflict)
		} else {
			res.Header().Set("Content-Type", "application/json")
			res.WriteHeader(http.StatusCreated)
		}

		response := models.APIShortenResponse{
			Result: addr + "/" + shortenURL,
		}

		enc := json.NewEncoder(res)
		if err := enc.Encode(response); err != nil {
			logger.Log.Error("error encoding response", zap.Error(err))
			http.Error(res, "Error adding new shorten URL", http.StatusBadRequest)
			return
		}
	}
}

func EncodeBatch(db Storage, addr string) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		var request []models.APIBatchRequest
		dec := json.NewDecoder(req.Body)
		if err := dec.Decode(&request); err != nil {
			logger.Log.Debug("can't decode request JSON body", zap.Error(err))
			http.Error(res, "Error adding new shorten URL", http.StatusBadRequest)
			return
		}

		var response []models.APIBatchResponse
		for _, url := range request {
			originalURL := url.OriginalURL
			if originalURL == "" {
				continue
			}

			shortenURL := base62.Base62Encode(rand.Uint64())
			ctx, cancel := context.WithTimeout(req.Context(), 3*time.Second)
			defer cancel()
			for !db.IsShortenUnique(ctx, shortenURL) {
				shortenURL = base62.Base62Encode(rand.Uint64())
			}

			ctx, cancel2 := context.WithTimeout(req.Context(), 3*time.Second)
			defer cancel2()
			err := db.AddURL(ctx, originalURL, shortenURL)
			if err != nil {
				http.Error(res, "Error adding new shorten URL", http.StatusBadRequest)
				return
			}
			response = append(response, models.APIBatchResponse{CorrelationID: url.CorrelationID, ShortenURL: addr + "/" + shortenURL})
		}

		res.Header().Set("Content-Type", "application/json")
		res.WriteHeader(http.StatusCreated)
		enc := json.NewEncoder(res)
		if err := enc.Encode(response); err != nil {
			logger.Log.Error("error encoding response", zap.Error(err))
			http.Error(res, "Error adding new shorten URL", http.StatusBadRequest)
			return
		}
	}
}

func CheckDBConnection(store Storage) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		db, ok := store.(*storage.Database)
		if !ok {
			http.Error(res, "Internal DB Error", http.StatusInternalServerError)
			return
		}

		ctx, cancel := context.WithTimeout(req.Context(), 1*time.Second)
		defer cancel()

		if err := db.DB.PingContext(ctx); err != nil {
			http.Error(res, "Error pinging DB", http.StatusInternalServerError)
			return
		}
		res.WriteHeader(http.StatusOK)
	}
}

func isUniqueViolationError(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
		return true
	}
	return false
}
