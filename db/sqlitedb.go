package db

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/mattn/go-sqlite3"
)

type SQLDB struct {
	Path              string
	BackupPath        string
	CacheSize         int
	sqldb             *sql.DB
	lastModifiedAt    int64
	lastModifiedAtWAL int64
	Modified          bool
}

func (db *SQLDB) Open(migrate string) error {
	var err error
	if db.CacheSize == 0 {
		db.CacheSize = 2000
	}

	err = mkParentDirs(db.Path)
	if err != nil {
		return fmt.Errorf("cannot create parent directories: %w", err)
	}

	err = mkParentDirs(db.BackupPath)
	if err != nil {
		return fmt.Errorf("cannot create parent directories: %w", err)
	}

	dsn := fmt.Sprintf("file:%s?_cache_size=%v&_journal_mode=WAL&_busy_timeout=5000", db.Path, -db.CacheSize)
	fmt.Println("Open", dsn)
	db.sqldb, err = sql.Open("sqlite3", dsn)
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

func (db *SQLDB) Query(query string, args ...any) (*sql.Rows, error) {
	return db.sqldb.Query(query, args...)
}

func (db *SQLDB) DBFileExists() bool {
	_, err := os.Stat(db.Path)
	if err == nil {
		// File exists, no need to download
		fmt.Println("db file exists", db.Path)
		return true
	} else if !os.IsNotExist(err) {
		// The error should be not exists
		// Some other error occurred
		fmt.Println("db file check Error", db.Path)
		return false
	}
	return false
}

// Returns the
func (db *SQLDB) UpdateModified() error {

	///
	// IF WAL HAS CHANGED FORCE UPDATE
	///

	walInfo, err := os.Stat(db.Path + "-wal")
	if err != nil {
		return err // Return error if file does not exist or cannot be accessed
	}
	walModifiedAt := walInfo.ModTime().Unix()
	if walModifiedAt != db.lastModifiedAtWAL {
		db.lastModifiedAtWAL = walModifiedAt
		// Try our best to force a checkpoint of the DB
		db.Exec("PRAGMA wal_checkpoint(FULL);")
	}

	//// Now onto checking the DB
	sqliteDBInfo, err := os.Stat(db.Path)
	if err != nil {
		return err // Return error if file does not exist or cannot be accessed
	}

	// Get the modification time and convert it to Unix timestamp
	lastModifiedAt := sqliteDBInfo.ModTime().Unix()

	//The lastModified time should be the greatest time of WAL and the db

	if lastModifiedAt != db.lastModifiedAt {
		db.lastModifiedAt = lastModifiedAt
		db.Modified = true
	} else {
		db.Modified = false
	}
	return nil
}

func (db *SQLDB) SqliteBackup() error {
	fmt.Println("SqliteBackup", db.Path, db.BackupPath)

	backupDB, err := sql.Open("sqlite3", db.BackupPath)
	if err != nil {
		return fmt.Errorf("cannot open backup db: %w", err)
	}
	defer backupDB.Close()

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

		backupConSQLite, ok := backupConRaw.(*sqlite3.SQLiteConn)
		if !ok {
			fmt.Println("SqliteBackup err backupConSQLite", db.Path, err)
			return fmt.Errorf("failed to get SQLiteConn backupConSQLite")
		}
		srcConn.Raw(func(srcConnRaw any) error {

			srcConnSQLite, ok := srcConnRaw.(*sqlite3.SQLiteConn)
			if !ok {
				fmt.Println("SqliteBackup err srcConnSQLite", db.Path, err)
				return fmt.Errorf("failed to get SQLiteConn srcConnSQLite")
			}

			backup, err := backupConSQLite.Backup("main", srcConnSQLite, "main")
			if err != nil {
				fmt.Println("SqliteBackup err Backup", db.Path, err)
				return fmt.Errorf("cannot create backup no create: %w", err)
			}

			defer backup.Finish()
			_, err = backup.Step(-1)

			if err != nil {
				fmt.Println("SqliteBackup err 3", db.Path, err)
				return fmt.Errorf("cannot create backup no step 1: %w", err)
			}

			return nil
		})
		return nil
	})
	if err != nil {
		fmt.Println("SqliteBackup Error on Raw", db.Path, err)
		return fmt.Errorf("cannot backup db: %w", err)
	}

	return nil
}

func trackTime(start time.Time, name string) {
	elapsed := time.Since(start)
	fmt.Printf("%s took %v\n", name, elapsed)
}

func mkParentDirs(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("cannot create directory %s: %w", dir, err)
	}
	return nil
}
