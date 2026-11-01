package backup_test

import (
	"database/sql"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
	"nighttroop.t1d.nz/utils/backup"
)

// A backup whose step fails still lets the next backup to the same path through
func TestSqliteBackupStepFails(t *testing.T) {
	dir := t.TempDir()
	open := func(dsn string) *sql.DB {
		db, err := sql.Open("sqlite", dsn)
		require.NoError(t, err)
		t.Cleanup(func() { db.Close() })
		require.NoError(t, db.Ping())
		return db
	}
	walDB := open("file:" + filepath.Join(dir, "wal.sqlite") + "?_journal_mode=WAL")
	path := filepath.Join(dir, "backup.sqlite")
	require.NoError(t, backup.SqliteBackup(t.Context(), walDB, path))

	// SQLite will not step a backup into a WAL DB of another page size
	smallPages := open("file:" + filepath.Join(dir, "small.sqlite") + "?_pragma=page_size(1024)")
	require.Error(t, backup.SqliteBackup(t.Context(), smallPages, path))

	require.NoError(t, backup.SqliteBackup(t.Context(), walDB, path))
}
