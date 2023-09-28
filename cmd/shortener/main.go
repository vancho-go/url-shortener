package main

import (
	"github.com/vancho-go/url-shortener/internal/app/helpers"
	"io"
	"math/rand"
	"net/http"
	url2 "net/url"
	"strings"
)

var dbInstance = make(map[string]string)

const addr = "localhost:8080"

func mainPage(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet && req.Method != http.MethodPost {
		http.Error(res, "Method not supported", http.StatusBadRequest)
		return
	}
	if req.Method == http.MethodGet {
		shortenURL := strings.TrimPrefix(req.URL.Path, "/")
		if originalURL, ok := dbInstance[shortenURL]; !ok {
			http.Error(res, "No such shorten URL", http.StatusBadRequest)
			return
		} else {
			res.Header().Set("Location", originalURL)
			res.WriteHeader(http.StatusTemporaryRedirect)
		}
	}

	if req.Method == http.MethodPost {
		originalURL, err := io.ReadAll(req.Body)
		if err != nil {
			_, _ = res.Write([]byte(err.Error()))
			return
		}
		if string(originalURL) == "" {
			http.Error(res, "URL parameter is missing", http.StatusBadRequest)
			return
		}
		id := helpers.Base62Encode(rand.Uint64())
		dbInstance[id] = string(originalURL)
		res.WriteHeader(http.StatusCreated)
		shortenURL, err := url2.JoinPath("http://", addr, id)
		if err != nil {
			_, _ = res.Write([]byte(err.Error()))
			return
		}
		_, _ = res.Write([]byte(shortenURL))
	}

}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc(`/`, mainPage)
	err := http.ListenAndServe(addr, mux)
	if err != nil {
		panic(err)
	}
}
