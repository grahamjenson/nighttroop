package db

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/mattn/go-sqlite3"
)

type SQLDB struct {
	Path           string
	BackupPath     string
	sqldb          *sql.DB
	lastModifiedAt int64
	Modified       bool
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

// Returns the
func (db *SQLDB) UpdateModified() error {
	fmt.Println("UpdateModified", db.Path)
	fileInfo, err := os.Stat(db.Path)
	if err != nil {
		return err // Return error if file does not exist or cannot be accessed
	}

	// Get the modification time and convert it to Unix timestamp
	lastModifiedAt := fileInfo.ModTime().Unix()

	if lastModifiedAt != db.lastModifiedAt {
		db.lastModifiedAt = lastModifiedAt
		db.Modified = true
		fmt.Println("UpdateModified MODIFIED", db.Path)
	} else {
		fmt.Println("UpdateModified NOT MODIFIED", db.Path)
		db.Modified = false
	}
	return nil
}

func (db *SQLDB) SqliteBackup() error {
	fmt.Println("SqliteBackup", db.Path)
	backupDB, err := sql.Open("sqlite3", db.BackupPath)
	if err != nil {
		return fmt.Errorf("cannot open backup db: %w", err)
	}
	defer backupDB.Close()
	err = backupDB.Ping()
	if err != nil {
		fmt.Println("Failed to connect to the destination database:", err)
	}

	backupCon, err := backupDB.Conn(context.TODO())
	if err != nil {
		return fmt.Errorf("cannot get backup connection: %w", err)
	}
	defer backupCon.Close()

	srcConn, err := db.sqldb.Conn(context.TODO())
	if err != nil {
		return fmt.Errorf("cannot get source connection: %w", err)
	}
	defer srcConn.Close()

	err = backupCon.Raw(func(backupConRaw any) error {
		fmt.Println("SqliteBackup dest RAW", db.Path)
		backupConSQLite, ok := backupConRaw.(*sqlite3.SQLiteConn)
		if !ok {
			return fmt.Errorf("failed to get SQLiteConn backupConSQLite")
		}
		srcConn.Raw(func(srcConnRaw any) error {
			fmt.Println("SqliteBackup src RAW 1", db.Path)
			srcConnSQLite, ok := srcConnRaw.(*sqlite3.SQLiteConn)
			if !ok {
				return fmt.Errorf("failed to get SQLiteConn srcConnSQLite")
			}
			fmt.Println("SqliteBackup src RAW 2", db.Path)
			backup, err := backupConSQLite.Backup("main", srcConnSQLite, "main")
			if err != nil {
				fmt.Println("SqliteBackup src RAW 2", db.Path, err)
				return fmt.Errorf("cannot create backup no create: %w", err)
			}
			fmt.Println("SqliteBackup src RAW 3", db.Path)
			defer backup.Finish()
			done, err := backup.Step(-1)
			fmt.Println("SqliteBackup src RAW 4", db.Path)
			if err != nil {
				return fmt.Errorf("cannot create backup no step 1: %w", err)
			}
			fmt.Println("SqliteBackup STEP 1", done)
			time.Sleep(2 * time.Second)
			done, err = backup.Step(-1)
			if err != nil {
				return fmt.Errorf("cannot create backup no step 2: %w", err)
			}
			fmt.Println("SqliteBackup STEP 2", done)
			return nil
		})
		return nil
	})
	if err != nil {
		fmt.Println("SqliteBackup Error on Raw", db.Path, err)
		return fmt.Errorf("cannot backup db: %w", err)
	}
	fmt.Println("SqliteBackup DONE", db.Path)
	return nil
}
