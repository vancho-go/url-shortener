package storage

import "errors"

type MapDB map[string]string

func (storage MapDB) AddURL(originalURL, shortenURL string) error {
	storage[shortenURL] = originalURL
	return nil
}

func (storage MapDB) GetURL(shortenURL string) (string, error) {
	originalURL, ok := storage[shortenURL]
	if !ok {
		return "", errors.New("no such shorten URL")
	}
	return originalURL, nil
}

func (storage MapDB) IsShortenUnique(shortenURL string) bool {
	_, ok := storage[shortenURL]
	return !ok
}
