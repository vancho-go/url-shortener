package handlers

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const addr = "localhost:8080"

func TestMainPage(t *testing.T) {
	var shortenURL string

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
			name:   "Test POST: missing URL error",
			method: http.MethodPost, reqBody: "",
			target: "/",
			want:   want{code: 400, response: "URL parameter is missing\n", contentType: "text/plain; charset=utf-8"},
		},
		{
			name:   "Test POST: created",
			method: http.MethodPost, reqBody: "https://practicum.yandex.ru",
			target: "/",
			want:   want{code: 201, contentType: ""},
		},
		{
			name:   "Test GET: existingURL",
			method: http.MethodGet, reqBody: "",
			want: want{code: 307, contentType: ""},
		},
		{
			name:   "Test GET: non-existingURL",
			method: http.MethodGet, reqBody: "https://practicum.yandex.ru",
			target: "/nonexist",
			want:   want{code: 400, response: "No such shorten URL\n", contentType: "text/plain; charset=utf-8"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "Test GET: existingURL" {
				tt.target = shortenURL
			}
			request := httptest.NewRequest(tt.method, tt.target, strings.NewReader(tt.reqBody))
			w := httptest.NewRecorder()
			hF := MainPage(addr)
			hF(w, request)

			res := w.Result()
			assert.Equal(t, tt.want.code, res.StatusCode)

			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)
			require.NoError(t, err)

			if res.StatusCode == 201 && string(resBody) != "" && tt.method == http.MethodPost {
				lastIndex := strings.LastIndex(string(resBody), "/")
				shortenURL = string(resBody)[lastIndex:]
				fmt.Println(shortenURL)
				assert.Equal(t, tt.want.contentType, res.Header.Get("Content-Type"))

			} else {
				assert.Equal(t, tt.want.contentType, res.Header.Get("Content-Type"))
				assert.Equal(t, tt.want.response, string(resBody))
			}
		})
	}
}
