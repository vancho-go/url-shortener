package storage

import "errors"

type Storage interface {
	AddURL(string, string) error
	GetURL(string) (string, error)
}

type DBInstance map[string]string

func (storage DBInstance) AddURL(originalURL, shortenURL string) error {
	storage[shortenURL] = originalURL
	return nil
}

func (storage DBInstance) GetURL(shortenURL string) (string, error) {
	originalURL, ok := storage[shortenURL]
	if !ok {
		return "", errors.New("no such shorten URL")
	}
	return originalURL, nil
}
