package storage

import (
	"context"
	"database/sql"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/vancho-go/url-shortener/internal/app/logger"
	"github.com/vancho-go/url-shortener/internal/app/models"
)

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
	selectQuery := "SELECT original_url FROM urls WHERE shorten_url=$1"
	stmt, err := db.DB.Prepare(selectQuery)
	if err != nil {
		return "", err
	}
	defer stmt.Close()

	row := stmt.QueryRowContext(ctx, shortenURL)

	var originalURL string
	err = row.Scan(&originalURL)
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
	if err != nil {
		return nil, err
	}
	var userURLs []models.APIUserURLResponse
	for rows.Next() {
		var userURL models.APIUserURLResponse
		err = rows.Scan(&userURL.ShortenURL, &userURL.OriginalURL)
		if err != nil {
			return nil, err
		}
		userURLs = append(userURLs, userURL)
	}
	return userURLs, nil
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
