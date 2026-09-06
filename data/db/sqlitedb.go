package db

import (
	"database/sql"
	"fmt"

	"nighttroop.t1d.nz/utils"
)

func OpenSqliteDB(path string, cacheSize int, inMemory bool) (*sql.DB, error) {
	var err error

	if cacheSize == 0 {
		cacheSize = 2000
	}

	if inMemory {
		sqldb, err := sql.Open("sqlite3", ":memory:")
		if err != nil {
			return nil, fmt.Errorf("cannot open sqlite db: %w", err)
		}

		sqldb.SetMaxOpenConns(1)
		sqldb.SetMaxIdleConns(1)

		return sqldb, nil
	}

	err = utils.MkParentDirs(path)
	if err != nil {
		return nil, fmt.Errorf("cannot create parent directories: %w", err)
	}

	dsn := fmt.Sprintf("file:%s?_cache_size=%v&_journal_mode=WAL&_busy_timeout=5000&_foreign_keys=1", path, -cacheSize)

	sqldb, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("cannot open sqlite db: %w", err)
	}

	return sqldb, nil
}
