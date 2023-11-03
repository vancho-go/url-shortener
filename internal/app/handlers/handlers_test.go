package handlers

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vancho-go/url-shortener/internal/app/storage"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const addr = "localhost:8080"

var dbInstance = make(storage.MapDB)

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
			handlerFunc := EncodeURL(dbInstance, addr)
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
			handlerFunc := DecodeURL(dbInstance)
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
			handlerFunc := EncodeURLJSON(dbInstance, addr)
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
