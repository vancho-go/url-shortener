// Модуль handlers сожержит логику обработчиков.
package http

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/vancho-go/url-shortener/internal/app/handlers/http/middlewares"
	"io"
	"math/rand"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/zap"

	"github.com/vancho-go/url-shortener/internal/app/base62"
	"github.com/vancho-go/url-shortener/internal/app/models"
	"github.com/vancho-go/url-shortener/internal/app/storage"
)

// DecodeURL возвращает оригинальный URL из хранилища для переданного сокращенного URL.
func DecodeURL(db storage.URLStorager) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		shortenURL := chi.URLParam(req, "shortenURL")
		ctx, cancel := context.WithTimeout(req.Context(), 1*time.Second)
		defer cancel()
		originalURL, err := db.GetURL(ctx, shortenURL)
		if err == nil {
			res.Header().Set("Location", originalURL)
			res.WriteHeader(http.StatusTemporaryRedirect)
			return
		}

		if errors.Is(err, storage.ErrDeletedURL) {
			res.WriteHeader(http.StatusGone)
			return
		}
		http.Error(res, "No such shorten URL", http.StatusBadRequest)
	}
}

// EncodeURL генерирует сокращенный URL для переданного оригинального URL.
func EncodeURL(db storage.URLStorager, addr string) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		cookie, err := getCookie(req)
		if err != nil {
			middlewares.Log.Debug("no cookie in request, got from context")
		}

		var userID string
		if cookie != nil {
			userID, err = middlewares.GetUserID(cookie.Value)
			if err != nil {
				middlewares.Log.Warn("something wrong with user_id")
			}
		}

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
		ctx, cancel := context.WithTimeout(req.Context(), 1*time.Second)
		defer cancel()
		for !db.IsShortenUnique(ctx, shortenURL) {
			shortenURL = base62.Base62Encode(rand.Uint64())
		}

		err = db.AddURL(ctx, string(originalURL), shortenURL, userID)
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
			middlewares.Log.Error("write failed", zap.Error(err))
		}
	}
}

// EncodeURLJSON генерирует сокращенный URL для переданного оригинального URL (в json).
func EncodeURLJSON(db storage.URLStorager, addr string) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		cookie, err := getCookie(req)
		if err != nil {
			middlewares.Log.Debug("no cookie in request, got from context")
		}

		var userID string
		if cookie != nil {
			userID, err = middlewares.GetUserID(cookie.Value)
			if err != nil {
				middlewares.Log.Warn("something wrong with user_id")
			}
		}

		var request models.APIShortenRequest
		dec := json.NewDecoder(req.Body)
		if err = dec.Decode(&request); err != nil {
			middlewares.Log.Warn("can't decode request JSON body", zap.Error(err))
			http.Error(res, "Error adding new shorten URL", http.StatusBadRequest)
			return
		}

		originalURL := request.URL
		if originalURL == "" {
			http.Error(res, "URL parameter is missing", http.StatusBadRequest)
			return
		}

		shortenURL := base62.Base62Encode(rand.Uint64())
		ctx, cancel := context.WithTimeout(req.Context(), 1*time.Second)
		defer cancel()
		for !db.IsShortenUnique(ctx, shortenURL) {
			shortenURL = base62.Base62Encode(rand.Uint64())
		}

		err = db.AddURL(ctx, originalURL, shortenURL, userID)
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
			middlewares.Log.Error("error encoding response", zap.Error(err))
			http.Error(res, "Error adding new shorten URL", http.StatusBadRequest)
			return
		}
	}
}

// EncodeBatch batch сокращенных URL для batch оригинальных URL.
func EncodeBatch(db storage.URLStorager, addr string) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		cookie, err := getCookie(req)
		if err != nil {
			middlewares.Log.Debug("no cookie in request, got from context")
		}

		var userID string
		if cookie != nil {
			userID, err = middlewares.GetUserID(cookie.Value)
			if err != nil {
				middlewares.Log.Warn("something wrong with user_id")
			}
		}

		var request []models.APIBatchRequest
		dec := json.NewDecoder(req.Body)
		if err := dec.Decode(&request); err != nil {
			middlewares.Log.Warn("can't decode request JSON body", zap.Error(err))
			http.Error(res, "Error adding new shorten URL", http.StatusBadRequest)
			return
		}

		var batch []models.APIBatchRequest
		var response []models.APIBatchResponse
		const batchSize = 100

		ctx, cancel := context.WithTimeout(req.Context(), 3*time.Second)
		defer cancel()

		for i, url := range request {
			originalURL := url.OriginalURL
			if originalURL == "" {
				continue
			}

			shortenURL := base62.Base62Encode(rand.Uint64())
			for !db.IsShortenUnique(ctx, shortenURL) {
				shortenURL = base62.Base62Encode(rand.Uint64())
			}

			batch = append(batch, models.APIBatchRequest{
				CorrelationID: url.CorrelationID,
				OriginalURL:   originalURL,
				ShortenURL:    shortenURL,
			})

			if len(batch) == batchSize || i == len(request)-1 {
				err := db.AddURLs(ctx, userID, batch...)
				if err != nil {
					http.Error(res, "Error adding new shorten URLs", http.StatusBadRequest)
					return
				}
				for _, b := range batch {
					response = append(response, models.APIBatchResponse{CorrelationID: b.CorrelationID, ShortenURL: addr + "/" + b.ShortenURL})
				}
				batch = nil // Сбросить пакет после вставки.
			}
		}

		res.Header().Set("Content-Type", "application/json")
		res.WriteHeader(http.StatusCreated)
		enc := json.NewEncoder(res)
		if err := enc.Encode(response); err != nil {
			middlewares.Log.Error("error encoding response", zap.Error(err))
			http.Error(res, "Error adding new shorten URL", http.StatusBadRequest)
			return
		}
	}
}

// GetUserURLs возвращает пользователю его ранее сокращенные URL.
func GetUserURLs(db storage.UserStorager, addr string) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		cookie, err := req.Cookie("AuthToken")
		if err != nil {
			middlewares.Log.Debug("error getting cookie", zap.Error(err))

			// Replace it after tests repaired
			http.Error(res, "No cookie presented", http.StatusUnauthorized)
			return
		}
		userID, err := middlewares.GetUserID(cookie.Value)
		if err != nil {
			middlewares.Log.Warn("something wrong with user_id", zap.Error(err))
			http.Error(res, "Bad user_id", http.StatusUnauthorized)
			return
		}
		ctx, cancel := context.WithTimeout(req.Context(), 1*time.Second)
		defer cancel()
		userURLs, err := db.GetUserURLs(ctx, userID)

		if len(userURLs) == 0 {
			res.WriteHeader(http.StatusNoContent)
			return
		}

		for i := 0; i < len(userURLs); i++ {
			userURLs[i].ShortenURL = addr + "/" + userURLs[i].ShortenURL
		}

		if err != nil {
			middlewares.Log.Error("error getting user urls", zap.Error(err))
			http.Error(res, "Error getting urls", http.StatusBadRequest)
			return
		}

		res.Header().Set("Content-Type", "application/json")
		res.WriteHeader(http.StatusOK)
		enc := json.NewEncoder(res)
		if err := enc.Encode(userURLs); err != nil {
			middlewares.Log.Error("error encoding response", zap.Error(err))
			http.Error(res, "Error adding new shorten URL", http.StatusBadRequest)
			return
		}
	}
}

// DeleteURLs удаляет URL пользователя.
func DeleteURLs(db storage.UserStorager) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		cookie, err := req.Cookie("AuthToken")
		if err != nil {
			middlewares.Log.Debug("error getting cookie", zap.Error(err))

			// Replace it after tests repaired
			http.Error(res, "No cookie presented", http.StatusNoContent)
			return
		}
		userID, err := middlewares.GetUserID(cookie.Value)
		if err != nil {
			middlewares.Log.Warn("something wrong with user_id", zap.Error(err))
			http.Error(res, "Bad user_id", http.StatusUnauthorized)
			return
		}

		var shortenUrls []string
		dec := json.NewDecoder(req.Body)
		if err = dec.Decode(&shortenUrls); err != nil {
			middlewares.Log.Warn("can't decode request JSON body", zap.Error(err))
			http.Error(res, "Error deleting shorten URLs", http.StatusBadRequest)
			return
		}
		urlsToDelete := make([]models.DeleteURLRequest, len(shortenUrls))
		for pos, url := range shortenUrls {
			urlsToDelete[pos].ShortenURL = url
			urlsToDelete[pos].UserID = userID
		}

		res.WriteHeader(http.StatusAccepted)

		ctx, cancel := context.WithTimeout(req.Context(), 60*time.Second)
		defer cancel()

		err = db.DeleteUserURLs(ctx, urlsToDelete...)
		if err != nil {
			middlewares.Log.Error("error deleting", zap.Error(err))
		}
	}
}

// CheckDBConnection пингует БД для проверки на доступность.
func CheckDBConnection(store storage.URLStorager) http.HandlerFunc {
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

// GetStats возвращает статистику по хранилищу.
func GetStats(store storage.StatsStorager) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		response, err := store.GetStats(req.Context())
		if err != nil {
			http.Error(res, "Error getting stats", http.StatusInternalServerError)
			return
		}
		res.Header().Set("Content-Type", "application/json")
		res.WriteHeader(http.StatusOK)
		enc := json.NewEncoder(res)
		if err := enc.Encode(response); err != nil {
			middlewares.Log.Error("error encoding response", zap.Error(err))
			http.Error(res, "Error getting stats", http.StatusBadRequest)
			return
		}
	}
}

// isUniqueViolationError проверяет является ли ошибка UniqueViolation.
func isUniqueViolationError(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation
}

// getCookie возвращает cookie пользователя.
func getCookie(req *http.Request) (*http.Cookie, error) {
	cookie, err := req.Cookie("AuthToken")

	if cookie == nil {
		cookie2, ok := req.Context().Value(middlewares.CookieKey).(*http.Cookie)
		if !ok {
			middlewares.Log.Debug("error conversion cookie")
		} else {
			cookie = cookie2
		}

	}
	return cookie, err
}
