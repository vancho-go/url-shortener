package handlers

import (
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/vancho-go/url-shortener/internal/app/models"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const addr = "localhost:8080"

//var dbInstance = make(storage.MapDB)

// MockStorager - это поддельная реализация Storager, используемая как в примере, так и в тестах.
type MockStorager struct {
	IsUniqueFunc func(ctx context.Context, shortenURL string) bool
	AddURLFunc   func(ctx context.Context, originalURL string, shortenURL string, userID string) error
	GetURLFunc   func(ctx context.Context, shortenURL string) (string, error)
}

func (m *MockStorager) AddURLs(ctx context.Context, requests []models.APIBatchRequest, s string) error {
	return nil
}

func (m *MockStorager) Close() error {
	return nil
}

func (m *MockStorager) GetUserURLs(ctx context.Context, s string) ([]models.APIUserURLResponse, error) {
	return []models.APIUserURLResponse{
		{
			ShortenURL:  "localhost:8080/qwerty22423",
			OriginalURL: "vk.com",
		},
		{
			ShortenURL:  "localhost:8080/kslfvk4",
			OriginalURL: "ya.ru",
		},
	}, nil
}

func (m *MockStorager) DeleteUserURLs(ctx context.Context, requests []models.DeleteURLRequest) error {
	return nil
}

func (m *MockStorager) IsShortenUnique(ctx context.Context, shortenURL string) bool {
	if m.IsUniqueFunc != nil {
		return m.IsUniqueFunc(ctx, shortenURL)
	}
	return true // Значение по умолчанию, если функция не задана
}

func (m *MockStorager) AddURL(ctx context.Context, originalURL string, shortenURL string, userID string) error {
	if m.AddURLFunc != nil {
		return m.AddURLFunc(ctx, originalURL, shortenURL, userID)
	}
	return nil // Значение по умолчанию, если функция не задана
}

func (m *MockStorager) GetURL(ctx context.Context, shortenURL string) (string, error) {
	if m.GetURLFunc != nil {
		return m.GetURLFunc(ctx, shortenURL)
	}
	if shortenURL == "48fnuid2" {
		return "http://example.com/encode", nil
	}
	return "", errors.New("not found") // Значение по умолчанию, если функция не задана
}

func TestEncodeURL(t *testing.T) {
	type want struct {
		code        int
		response    string
		contentType string
	}
	tests := []struct {
		name    string
		method  string
		reqBody string
		target  string
		want    want
	}{
		{
			name:    "Test POST: missing URL error",
			method:  http.MethodPost,
			reqBody: "",
			target:  "/",
			want:    want{code: 400, response: "URL parameter is missing\n", contentType: "text/plain; charset=utf-8"},
		},
		{
			name:    "Test POST: created",
			method:  http.MethodPost,
			reqBody: "https://practicum.yandex.ru",
			target:  "/",
			want:    want{code: 201, contentType: ""},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(tt.method, tt.target, strings.NewReader(tt.reqBody))
			w := httptest.NewRecorder()
			handlerFunc := EncodeURL(&MockStorager{}, addr)
			handlerFunc(w, request)

			res := w.Result()
			assert.Equal(t, tt.want.code, res.StatusCode)

			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)
			require.NoError(t, err)

			if res.StatusCode == 201 && string(resBody) != "" {
				assert.Equal(t, tt.want.contentType, res.Header.Get("Content-Type"))
			} else {
				assert.Equal(t, tt.want.contentType, res.Header.Get("Content-Type"))
				assert.Equal(t, tt.want.response, string(resBody))
			}
		})
	}
}

func TestDecodeURL(t *testing.T) {
	type want struct {
		code        int
		response    string
		contentType string
	}
	tests := []struct {
		name    string
		method  string
		reqBody string
		target  string
		want    want
	}{
		{
			name:   "Test GET: non-existingURL",
			method: http.MethodGet, reqBody: "https://practicum.yandex.ru",
			target: "/nonexist",
			want:   want{code: 400, response: "No such shorten URL\n", contentType: "text/plain; charset=utf-8"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(tt.method, tt.target, strings.NewReader(tt.reqBody))
			w := httptest.NewRecorder()
			handlerFunc := DecodeURL(&MockStorager{})
			handlerFunc(w, request)

			res := w.Result()
			assert.Equal(t, tt.want.code, res.StatusCode)

			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)
			require.NoError(t, err)

			assert.Equal(t, tt.want.contentType, res.Header.Get("Content-Type"))
			assert.Equal(t, tt.want.response, string(resBody))

		})
	}
}

func TestEncodeURLJSON(t *testing.T) {
	type want struct {
		code        int
		response    string
		contentType string
	}
	tests := []struct {
		name    string
		method  string
		reqBody string
		target  string
		want    want
	}{
		{
			name:    "Test POST: empty body",
			method:  http.MethodPost,
			reqBody: "",
			target:  "/api/shorten",
			want:    want{code: 400, response: "Error adding new shorten URL\n", contentType: "text/plain; charset=utf-8"},
		},
		{
			name:    "Test POST: empty URL field in JSON in body",
			method:  http.MethodPost,
			reqBody: `{"url": ""}`,
			target:  "/api/shorten",
			want:    want{code: 400, response: "URL parameter is missing\n", contentType: "text/plain; charset=utf-8"},
		},
		{
			name:    "Test POST: no URL field in JSON in body",
			method:  http.MethodPost,
			reqBody: `{"url2": ""}`,
			target:  "/api/shorten",
			want:    want{code: 400, response: "URL parameter is missing\n", contentType: "text/plain; charset=utf-8"},
		},
		{
			name:    "Test POST: created",
			method:  http.MethodPost,
			reqBody: `{"url": "vk.com"}`,
			target:  "/api/shorten",
			want:    want{code: 201, contentType: "application/json"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(tt.method, tt.target, strings.NewReader(tt.reqBody))
			w := httptest.NewRecorder()
			handlerFunc := EncodeURLJSON(&MockStorager{}, addr)
			handlerFunc(w, request)

			res := w.Result()
			assert.Equal(t, tt.want.code, res.StatusCode)

			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)
			require.NoError(t, err)

			if res.StatusCode == 201 && string(resBody) != "" {
				assert.Equal(t, tt.want.contentType, res.Header.Get("Content-Type"))
			} else {
				assert.Equal(t, tt.want.contentType, res.Header.Get("Content-Type"))
				assert.Equal(t, tt.want.response, string(resBody))
			}
		})
	}
}

func ExampleEncodeURL() {
	// Создаем MockStorager.
	db := MockStorager{}

	// Создаем HTTP-запрос с телом, содержащим оригинальный URL.
	originalURL := "http://example.com/original"
	req := httptest.NewRequest("POST", "http://localhost:8080", strings.NewReader(originalURL))

	// Создаем ResponseRecorder для записи ответа.
	rr := httptest.NewRecorder()

	// Создаем хендлер с использованием нашего MockStorager и адреса для сокращенных URL.
	handlerFunc := EncodeURL(&db, "http://localhost:8080")
	handlerFunc(rr, req)

	res := rr.Result()
	// Выводим результат.
	fmt.Printf("Status Code: %d\n", res.StatusCode)

	// Output:
	// Status Code: 201
}

func ExampleDecodeURL() {
	// Создаем MockStorager.
	db := MockStorager{}

	// Создаем роутер chi и регистрируем хендлер.
	r := chi.NewRouter()
	r.Get("/{shortenURL}", DecodeURL(&db))

	// Создаем тестовый сервер.
	ts := httptest.NewServer(r)
	defer ts.Close()

	// Создаем HTTP-запрос к тестовому серверу, содержащим существующий сокращенный URL.
	req := httptest.NewRequest("GET", ts.URL+"/48fnuid2", nil)

	// Создаем ResponseRecorder для записи ответа.
	rr := httptest.NewRecorder()

	// Делаем запрос через роутер chi, чтобы параметры URL были извлечены корректно.
	r.ServeHTTP(rr, req)

	res := rr.Result()
	fmt.Printf("Status Code: %d\n", res.StatusCode)

	// Output:
	// Status Code: 307
}

func ExampleEncodeURLJSON() {
	// Создаем MockStorager.
	db := MockStorager{}

	// Создаем HTTP-запрос с телом, содержащим оригинальный URL.
	req := httptest.NewRequest(
		"POST",
		"http://localhost:8080/api/shorten",
		strings.NewReader("{\"url\": \"vk.com\"}"))

	// Создаем ResponseRecorder для записи ответа.
	rr := httptest.NewRecorder()

	// Создаем хендлер с использованием нашего MockStorager и адреса для сокращенных URL.
	handlerFunc := EncodeURLJSON(&db, "localhost:8080")
	handlerFunc(rr, req)

	res := rr.Result()
	fmt.Printf("Status Code: %d\n", res.StatusCode)

	// Output:
	// Status Code: 201
}

func ExampleEncodeBatch() {
	// Создаем MockStorager.
	db := MockStorager{}

	// Создаем HTTP-запрос с телом, содержащим несколько оригинальных URL.
	req := httptest.NewRequest(
		"POST",
		"http://localhost:8080/api/shorten/batch",
		strings.NewReader("[{\"correlation_id\": \"ddd\",\"original_url\": \"facebook.com\"},{\"correlation_id\": \"ddd\",\"original_url\": \"youtube.com\"}]"))

	// Создаем ResponseRecorder для записи ответа.
	rr := httptest.NewRecorder()

	// Создаем хендлер с использованием нашего MockStorager и адреса для сокращенных URL.
	handlerFunc := EncodeBatch(&db, "localhost:8080")
	handlerFunc(rr, req)

	res := rr.Result()
	fmt.Printf("Status Code: %d\n", res.StatusCode)

	// Output:
	// Status Code: 201
}

func ExampleGetUserURLs() {
	// Создаем MockStorager.
	db := MockStorager{}

	// Создаем HTTP-запрос.
	req := httptest.NewRequest(
		"GET",
		"http://localhost:8080/api/user/urls",
		nil)

	// Создаем ResponseRecorder для записи ответа.
	rr := httptest.NewRecorder()

	// Создаем хендлер с использованием нашего MockStorager и адреса для сокращенных URL.
	handlerFunc := GetUserURLs(&db, "localhost:8080")
	handlerFunc(rr, req)

	res := rr.Result()
	fmt.Printf("Status Code: %d\n", res.StatusCode)

	// Output:
	// Status Code: 204
}

func ExampleDeleteURLs() {
	// Создаем MockStorager.
	db := MockStorager{}

	// Создаем HTTP-запрос.
	req := httptest.NewRequest(
		"DELETE",
		"http://localhost:8080/api/user/urls",
		nil)

	// Создаем ResponseRecorder для записи ответа.
	rr := httptest.NewRecorder()

	// Создаем хендлер с использованием нашего MockStorager и адреса для сокращенных URL.
	handlerFunc := DeleteURLs(&db)
	handlerFunc(rr, req)

	res := rr.Result()
	fmt.Printf("Status Code: %d\n", res.StatusCode)

	// Output:
	// Status Code: 204
}
