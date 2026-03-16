package db

import (
	"database/sql"
	"fmt"
	"os"

	"nighttroop.t1d.nz/utils"
)

func OpenSqliteDB(path string, cacheSize int) (*sql.DB, error) {
	var err error

	if cacheSize == 0 {
		cacheSize = 2000
	}

	err = utils.MkParentDirs(path)
	if err != nil {
		return nil, fmt.Errorf("cannot create parent directories: %w", err)
	}

	dsn := fmt.Sprintf("file:%s?_cache_size=%v&_journal_mode=WAL&_busy_timeout=5000", path, -cachesize)
	fmt.Println("Open", dsn)
	sqldb, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("cannot open sqlite db: %w", err)
	}

	return sqldb, nil
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
	return utils.FileExists(db.Path)
}

func (db *SQLDB) DBBackupFileExists() bool {
	return utils.FileExists(db.BackupPath)
}

// Returns the
func (db *SQLDB) UpdateModified() (bool, error) {

	///
	// FORCE UPDATE WAL UPDATE
	///

	db.Exec("PRAGMA wal_checkpoint(FULL);")

	//// Now onto checking the DB
	sqliteDBInfo, err := os.Stat(db.Path)
	if err != nil {
		return false, err // Return error if file does not exist or cannot be accessed
	}

	// Get the modification time and convert it to Unix timestamp
	lastModifiedAt := sqliteDBInfo.ModTime().Unix()

	//The lastModified time should be the greatest time of WAL and the db

	modified := false
	if lastModifiedAt != db.lastModifiedAt {
		db.lastModifiedAt = lastModifiedAt
		modified = true
	}

	if modified || !db.DBBackupFileExists() {
		err = db.SqliteBackup()
		if err != nil {
			return modified, fmt.Errorf("cannot backup user db: %w %v", err, db.Path)
		}
	}

	return modified, nil
}
