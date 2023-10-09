package storage

import "errors"

type MapDBInstance map[string]string

func (storage MapDBInstance) AddURL(originalURL, shortenURL string) error {
	storage[shortenURL] = originalURL
	return nil
}

func (storage MapDBInstance) GetURL(shortenURL string) (string, error) {
	originalURL, ok := storage[shortenURL]
	if !ok {
		return "", errors.New("no such shorten URL")
	}
	return originalURL, nil
}
