package storage

import (
	"encoding/json"
	"errors"
	"os"
	"sync"
)

type Data struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

type EncoderDecoder struct {
	file    *os.File
	storage map[string]string
	encoder *json.Encoder
	decoder *json.Decoder
	mu      sync.Mutex
}

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
	}, nil
}

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

func (ed *EncoderDecoder) Close() error {
	return ed.file.Close()
}

func (ed *EncoderDecoder) AddURL(originalURL, shortenURL string) error {
	data := &Data{ShortURL: shortenURL, OriginalURL: originalURL}
	ed.mu.Lock()
	ed.storage[shortenURL] = originalURL
	defer ed.mu.Unlock()
	return ed.encoder.Encode(&data)
}

func (ed *EncoderDecoder) GetURL(shortenURL string) (string, error) {
	ed.mu.Lock()
	originalURL, ok := ed.storage[shortenURL]
	defer ed.mu.Unlock()
	if !ok {
		return "", errors.New("no such shorten URL")
	}
	return originalURL, nil
}

func (ed *EncoderDecoder) IsShortenUnique(shortenURL string) bool {
	_, ok := ed.storage[shortenURL]
	return !ok
}
