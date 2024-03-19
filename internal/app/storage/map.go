package storage

import (
	"context"
	"errors"

	"github.com/vancho-go/url-shortener/internal/app/models"
)

// MapDB - key-value хранилище для URL.
type MapDB map[string]string

// AddURL сохраняет оригинальный и сокращенный URL в хранилище.
func (storage MapDB) AddURL(ctx context.Context, originalURL, shortenURL, userID string) error {
	storage[shortenURL] = originalURL
	return nil
}

// GetURL извлекает сокращенный URL для переданного оригинального URL из хранилища.
func (storage MapDB) GetURL(ctx context.Context, shortenURL string) (string, error) {
	originalURL, ok := storage[shortenURL]
	if !ok {
		return "", errors.New("no such shorten URL")
	}
	return originalURL, nil
}

// IsShortenUnique проверяет сокращенный URL на уникальность.
func (storage MapDB) IsShortenUnique(ctx context.Context, shortenURL string) bool {
	_, ok := storage[shortenURL]
	return !ok
}

// GetUserURLs извлекает URL из хранилища для конкретного пользователя.
func (storage MapDB) GetUserURLs(ctx context.Context, userID string) ([]models.APIUserURLResponse, error) {
	return nil, errors.New("method not implemented for this type of storage")
}

// DeleteUserURLs удаляет URL из хранилища для конкретного пользователя.
func (storage MapDB) DeleteUserURLs(ctx context.Context, urlsToDelete ...models.DeleteURLRequest) error {
	return errors.New("method not implemented for this type of storage")
}

// AddURLs сохраняет batch оригинальных и сокращенных URL в хранилище.
func (storage MapDB) AddURLs(ctx context.Context, userID string, urls ...models.APIBatchRequest) error {
	return errors.New("method not implemented for this type of storage")
}

// GetStats извлекает статистику хранилища.
func (storage MapDB) GetStats(ctx context.Context) (*models.APIStatsResponse, error) {
	return nil, errors.New("method not implemented for this type of storage")
}

// Close закрывает хранилище.
func (storage MapDB) Close() error {
	return nil
}
