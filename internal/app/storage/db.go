package storage

import (
	"database/sql"
	_ "github.com/jackc/pgx/v5/stdlib"
)

//type Database struct {
//	Conn *sql.DB
//}

func Initialize(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	return db, nil
}
