package storage

import (
	"context"
	"database/sql"
	_ "github.com/jackc/pgx/v5/stdlib"
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
	return &Database{DB: db}, nil
}

func (db *Database) AddURL(string, string) error {
	return nil
}

func (db *Database) GetURL(string) (string, error) {
	return "", nil
}

func (db *Database) IsShortenUnique(string) bool {
	return false
}

func (db *Database) Close() error {
	return db.DB.Close()
}
