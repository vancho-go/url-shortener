package storage

import (
	"context"
	"database/sql"
	"errors"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/vancho-go/url-shortener/internal/app/logger"
	"github.com/vancho-go/url-shortener/internal/app/models"
	"go.uber.org/zap"
	"sync"
)

var ErrDeletedURL = errors.New("URL was deleted")

type Database struct {
	DB *sql.DB
}

func Initialize(dsn string) (*Database, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	err = CreateIfNotExists(db)
	if err != nil {
		return nil, err
	}
	return &Database{DB: db}, nil
}

func CreateIfNotExists(db *sql.DB) error {
	createTableQuery := `
		CREATE TABLE IF NOT EXISTS urls (
			id SERIAL PRIMARY KEY,
			user_id VARCHAR NOT NULL,
			shorten_url VARCHAR NOT NULL,
			original_url VARCHAR NOT NULL,
			deleted BOOLEAN DEFAULT FALSE NOT NULL,
			UNIQUE (shorten_url),
		    UNIQUE (original_url)
		);`

	_, err := db.Exec(createTableQuery)
	if err != nil {
		return err
	}
	return nil
}

func (db *Database) AddURL(ctx context.Context, originalURL, shortenURL, userID string) error {
	tx, err := db.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	insertQuery := "INSERT INTO urls (shorten_url, original_url, user_id) VALUES ($1, $2, $3)"
	stmt, err := db.DB.PrepareContext(ctx, insertQuery)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, shortenURL, originalURL, userID)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (db *Database) GetURL(ctx context.Context, shortenURL string) (string, error) {
	selectQuery := "SELECT original_url, deleted FROM urls WHERE shorten_url=$1"
	stmt, err := db.DB.Prepare(selectQuery)
	if err != nil {
		return "", err
	}
	defer stmt.Close()

	row := stmt.QueryRowContext(ctx, shortenURL)

	var originalURL string
	var deleted bool
	err = row.Scan(&originalURL, &deleted)
	if deleted {
		return "", ErrDeletedURL
	}
	if err != nil {
		return "", err
	}
	return originalURL, nil

}

func (db *Database) GetUserURLs(ctx context.Context, userID string) ([]models.APIUserURLResponse, error) {
	selectQuery := "SELECT shorten_url, original_url FROM urls WHERE user_id=$1"
	stmt, err := db.DB.Prepare(selectQuery)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.QueryContext(ctx, userID)

	if rows.Err() != nil {
		return nil, rows.Err()
	}
	defer rows.Close()

	if err != nil {
		return nil, err
	}

	var userURLs []models.APIUserURLResponse
	for rows.Next() {
		var userURL models.APIUserURLResponse
		err := rows.Scan(&userURL.ShortenURL, &userURL.OriginalURL)
		if err != nil {
			return nil, err
		}
		userURLs = append(userURLs, userURL)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return userURLs, nil
}

func (db *Database) DeleteUserURLs(ctx context.Context, urlsToDelete []models.DeleteURLRequest) error {
	// Получаем канал с данными
	inputCh := generateDeleteURLChan(ctx, urlsToDelete)

	// Отдаем канал с данными, генерируем 5 воркеров
	// которые будут делать запрос на удаление из БД
	// и получаем каналы ответов этих воркеров
	channels := fanOutDeleters(ctx, inputCh, db)

	// Отправляем полученные каналы ответов, чтобы их все обработать в одном месте
	deleteResCh := fanIn(ctx, channels...)

	for err := range deleteResCh {
		if err != nil {
			logger.Log.Error("error deleting row", zap.Error(err))
		}
	}
	return nil
}

// ex generator
func generateDeleteURLChan(ctx context.Context, input []models.DeleteURLRequest) chan models.DeleteURLRequest {
	inputCh := make(chan models.DeleteURLRequest)

	go func() {
		defer close(inputCh)

		for _, deleteURL := range input {
			select {
			case <-ctx.Done():
				return
			case inputCh <- deleteURL:
			}
		}
	}()

	return inputCh
}

func urlDeleter(ctx context.Context, inputCh chan models.DeleteURLRequest, db *Database) chan error {
	deleteRes := make(chan error)

	go func() {
		defer close(deleteRes)

		for url := range inputCh {
			err := db.deleteUserURL(ctx, url)

			select {
			case <-ctx.Done():
				return
			case deleteRes <- err:
			}
		}
	}()
	return deleteRes
}

func fanOutDeleters(ctx context.Context, inputCh chan models.DeleteURLRequest, db *Database) []chan error {
	numWorkers := 5
	channels := make([]chan error, numWorkers)

	for i := 0; i < numWorkers; i++ {
		deleteResCh := urlDeleter(ctx, inputCh, db)
		channels[i] = deleteResCh
	}
	return channels
}

func fanIn(ctx context.Context, resultChs ...chan error) chan error {
	finalCh := make(chan error)

	var wg sync.WaitGroup

	for _, ch := range resultChs {
		ch2 := ch
		wg.Add(1)

		go func() {
			defer wg.Done()

			for data := range ch2 {
				select {
				case <-ctx.Done():
					return
				case finalCh <- data:
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(finalCh)
	}()
	return finalCh
}

func (db *Database) deleteUserURL(ctx context.Context, urlToDelete models.DeleteURLRequest) error {
	tx, err := db.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare("UPDATE urls SET deleted = true WHERE user_id = $1 AND shorten_url = $2")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, urlToDelete.UserID, urlToDelete.ShortenURL)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

func (db *Database) GetShortenURLByOriginal(ctx context.Context, originalURL string) (string, error) {
	selectQuery := "SELECT shorten_url FROM urls WHERE original_url=$1"
	stmt, err := db.DB.Prepare(selectQuery)
	if err != nil {
		return "", err
	}
	defer stmt.Close()

	row := stmt.QueryRowContext(ctx, originalURL)

	var shortenURL string
	err = row.Scan(&shortenURL)
	if err != nil {
		return "", err
	}
	return shortenURL, nil
}

func (db *Database) IsShortenUnique(ctx context.Context, shortenURL string) bool {
	selectQuery := "SELECT COUNT(*) FROM urls WHERE shorten_url=$1"
	stmt, err := db.DB.Prepare(selectQuery)
	if err != nil {
		logger.Log.Error("error in preparing query for unique count")
		//TODO
		return false
	}
	defer stmt.Close()

	row := stmt.QueryRowContext(ctx, shortenURL)

	var count int
	err = row.Scan(&count)
	if err != nil {
		logger.Log.Error("error in scanning count query")
		//TODO
		return false
	}
	return count == 0
}

func (db *Database) Close() error {
	return db.DB.Close()
}
