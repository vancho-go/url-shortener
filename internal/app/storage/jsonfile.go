package storage

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"sync"

	"github.com/vancho-go/url-shortener/internal/app/models"
)

// Data хранит оигинальный и сокращенный URL.
type Data struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

// EncoderDecoder объект, реализующий интерфейс storage.
type EncoderDecoder struct {
	file    *os.File
	storage map[string]string
	encoder *json.Encoder
	decoder *json.Decoder
	mu      sync.Mutex
}

// NewEncoderDecoder конструктор EncoderDecoder объекта.
func NewEncoderDecoder(filename string) (*EncoderDecoder, error) {
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	return &EncoderDecoder{
		file:    file,
		storage: make(map[string]string),
		encoder: json.NewEncoder(file),
		decoder: json.NewDecoder(file),
		mu:      sync.Mutex{},
	}, nil
}

// Initialize создает хранилище и достает сохраненные сокращенные url из файла в память.
func (ed *EncoderDecoder) Initialize() error {
	decoder := ed.decoder
	for decoder.More() {
		var data Data
		err := decoder.Decode(&data)
		if err != nil {
			return err
		}
		ed.storage[data.ShortURL] = data.OriginalURL
	}
	return nil
}

// Close закрывает хранилище.
func (ed *EncoderDecoder) Close() error {
	return ed.file.Close()
}

// GetUserURLs извлекает URL из хранилища для конкретного пользователя.
func (ed *EncoderDecoder) GetUserURLs(ctx context.Context, userID string) ([]models.APIUserURLResponse, error) {
	return nil, errors.New("method not implemented for this type of storage")
}

// DeleteUserURLs удаляет URL из хранилища для конкретного пользователя.
func (ed *EncoderDecoder) DeleteUserURLs(ctx context.Context, urlsToDelete ...models.DeleteURLRequest) error {
	return errors.New("method not implemented for this type of storage")
}

// AddURLs сохраняет batch оригинальных и сокращенных URL в хранилище.
func (ed *EncoderDecoder) AddURLs(ctx context.Context, userID string, urls ...models.APIBatchRequest) error {
	return errors.New("method not implemented for this type of storage")
}

// AddURL сохраняет оригинальный и сокращенный URL в хранилище.
func (ed *EncoderDecoder) AddURL(ctx context.Context, originalURL, shortenURL, userID string) error {
	data := &Data{ShortURL: shortenURL, OriginalURL: originalURL}
	ed.mu.Lock()
	ed.storage[shortenURL] = originalURL
	defer ed.mu.Unlock()
	return ed.encoder.Encode(&data)
}

// GetURL извлекает сокращенный URL для переданного оригинального URL из хранилища.
func (ed *EncoderDecoder) GetURL(ctx context.Context, shortenURL string) (string, error) {
	ed.mu.Lock()
	originalURL, ok := ed.storage[shortenURL]
	defer ed.mu.Unlock()
	if !ok {
		return "", errors.New("no such shorten URL")
	}
	return originalURL, nil
}

// IsShortenUnique проверяет сокращенный URL на уникальность.
func (ed *EncoderDecoder) IsShortenUnique(ctx context.Context, shortenURL string) bool {
	_, ok := ed.storage[shortenURL]
	return !ok
}

// GetStats извлекает статистику хранилища.
func (ed *EncoderDecoder) GetStats(ctx context.Context) (*models.APIStatsResponse, error) {
	return nil, errors.New("method not implemented for this type of storage")
}
