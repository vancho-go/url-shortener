package handlers

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"net/http/httptest"
	"strings"
)

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
