package storage

import (
	"context"
	"errors"
)

type MapDB map[string]string

func (storage MapDB) AddURL(ctx context.Context, originalURL, shortenURL string) error {
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

func (storage MapDB) Close() error {
	return nil
}
