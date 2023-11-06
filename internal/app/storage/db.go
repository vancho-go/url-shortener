package storage

import (
	"context"
	"database/sql"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/vancho-go/url-shortener/internal/app/logger"
)

type Database struct {
	DB *sql.DB
}

func Initialize(ctx context.Context, dsn string) (*Database, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return &Database{}, err
	}

	err = db.PingContext(ctx)
	if err != nil {
		return &Database{}, err
	}

	err = CreateIfNotExists(db)
	if err != nil {
		return &Database{}, err
	}
	return &Database{DB: db}, nil
}

func CreateIfNotExists(db *sql.DB) error {
	createTableQuery := `
		CREATE TABLE IF NOT EXISTS urls (
			id SERIAL PRIMARY KEY,
			shorten_url VARCHAR NOT NULL,
			original_url VARCHAR NOT NULL,
			UNIQUE (shorten_url)
		);`

	_, err := db.Exec(createTableQuery)
	if err != nil {
		return err
	}
	return nil
}

func (db *Database) AddURL(ctx context.Context, originalURL, shortenURL string) error {
	insertQuery := "INSERT INTO urls (shorten_url, original_url) VALUES ($1, $2)"
	stmt, err := db.DB.Prepare(insertQuery)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, shortenURL, originalURL)
	if err != nil {
		return err
	}
	return nil
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
	_ = row.Scan(&count)
	//if err != nil {
	//	logger.Log.Error("error in scanning count query")
	//	//TODO
	//	return false
	//}
	return count == 0
}

func (db *Database) Close() error {
	return db.DB.Close()
}
