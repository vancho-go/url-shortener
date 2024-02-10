package storage

import (
	"context"
	"errors"
	"github.com/vancho-go/url-shortener/internal/app/config"
	"github.com/vancho-go/url-shortener/internal/app/logger"
	"github.com/vancho-go/url-shortener/internal/app/models"
)

// URLStorager реализует методы для работы с URL.
type URLStorager interface {
	// AddURL сохраняет оригинальный и сокращенный URL в хранилище.
	AddURL(context.Context, string, string, string) error
	// AddURLs сохраняет batch оригинальных и сокращенных URL в хранилище.
	AddURLs(context.Context, string, ...models.APIBatchRequest) error
	// GetURL извлекает сокращенный URL для переданного оригинального URL из хранилища.
	GetURL(context.Context, string) (string, error)
	// IsShortenUnique проверяет сокращенный URL на уникальность.
	IsShortenUnique(context.Context, string) bool
	// Close закрывает хранилище.
	Close() error
}

// UserStorager реализует методы для работы с пользователями.
type UserStorager interface {
	// GetUserURLs извлекает URL из хранилища для конкретного пользователя.
	GetUserURLs(context.Context, string) ([]models.APIUserURLResponse, error)
	// DeleteUserURLs удаляет URL из хранилища для конкретного пользователя.
	DeleteUserURLs(context.Context, ...models.DeleteURLRequest) error
}

type Storager interface {
	URLStorager
	UserStorager
}

func New(serverConfig config.ServerConfig) (Storager, error) {
	switch {
	case serverConfig.DBDSN != "":
		logger.Log.Info("Initializing postgres storage")
		db, err := Initialize(serverConfig.DBDSN)
		if err != nil {
			return nil, errors.New("error Postgres DB initializing")
		}
		return db, nil

	case serverConfig.FileStorage != "":
		logger.Log.Info("Initializing file storage")
		dbInstance, err := NewEncoderDecoder(serverConfig.FileStorage)
		if err != nil {
			return nil, errors.New("error in FileStorage constructor")
		}

		err = dbInstance.Initialize()
		if err != nil {
			return nil, errors.New("error in FileStorage initializing")
		}
		return dbInstance, nil

	default:
		logger.Log.Info("Initializing in-memory storage")
		return MapDB{}, nil
	}
}
