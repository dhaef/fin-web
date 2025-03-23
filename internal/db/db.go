package db

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

func NewDbConnection(path string) (*sql.DB, error) {
	conn, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	return conn, nil
}
