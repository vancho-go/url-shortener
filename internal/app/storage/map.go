package storage

import (
	"context"
	"errors"

	"github.com/vancho-go/url-shortener/internal/app/models"
)

type MapDB map[string]string

func (storage MapDB) AddURL(ctx context.Context, originalURL, shortenURL, userID string) error {
	storage[shortenURL] = originalURL
	return nil
}

func (storage MapDB) GetURL(ctx context.Context, shortenURL string) (string, error) {
	originalURL, ok := storage[shortenURL]
	if !ok {
		return "", errors.New("no such shorten URL")
	}
	return originalURL, nil
}

func (storage MapDB) IsShortenUnique(ctx context.Context, shortenURL string) bool {
	_, ok := storage[shortenURL]
	return !ok
}
func (storage MapDB) GetUserURLs(ctx context.Context, userID string) ([]models.APIUserURLResponse, error) {
	return nil, errors.New("method not implemented for this type of storage")
}

func (storage MapDB) DeleteUserURLs(ctx context.Context, urlsToDelete []models.DeleteURLRequest) error {
	return errors.New("method not implemented for this type of storage")
}

func (storage MapDB) AddURLs(ctx context.Context, urls []models.APIBatchRequest, userID string) error {
	return errors.New("method not implemented for this type of storage")
}

func (storage MapDB) Close() error {
	return nil
}
