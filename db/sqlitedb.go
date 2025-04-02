package db

import (
	"database/sql"
	"fmt"
)

type SQLDB struct {
	Path  string
	sqldb *sql.DB
}

func (db *SQLDB) Open(migrate string) error {
	var err error

	db.sqldb, err = sql.Open("sqlite3", db.Path)
	if err != nil {
		return err
	}

	// MIGRATE
	// Create table for storing page views.
	if _, err := db.Exec(migrate); err != nil {
		return fmt.Errorf("cannot create table: %w", err)
	}

	return nil
}

func (db *SQLDB) Close() error {
	if db.sqldb != nil {
		return db.sqldb.Close()
	}
	return nil
}

func (db *SQLDB) Exec(query string, args ...any) (sql.Result, error) {
	return db.sqldb.Exec(query, args...)
}
