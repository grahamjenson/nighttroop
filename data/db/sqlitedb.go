package db

import (
	"database/sql"
	"fmt"

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

	dsn := fmt.Sprintf("file:%s?_cache_size=%v&_journal_mode=WAL&_busy_timeout=5000", path, -cacheSize)

	sqldb, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("cannot open sqlite db: %w", err)
	}

	return sqldb, nil
}
