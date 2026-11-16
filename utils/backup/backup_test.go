package backup_test

import (
	"database/sql"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/pierrec/lz4/v4"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/utils/backup"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/testutils"
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

// readings is how many glucose readings the DB at path has, opened by itself, without the write-ahead log of the DB it
// was copied from
func readings(t *testing.T, path string) int {
	db, err := sitedb.OpenSqliteDB(path, 0)
	require.NoError(t, err)
	defer db.Close()
	got, err := sitedb.New(db).SearchGlucoseReadings(t.Context(), sitedb.SearchGlucoseReadingsParams{Limit: 1000})
	require.NoError(t, err)
	return len(got)
}

// A site's backup is compressed with LZ4 and has every write, also those only in the write-ahead log, and leaves no
// temp file beside it
func TestCompressedBackupHasWhatIsOnlyInTheWAL(t *testing.T) {
	cfg := *testutils.GetTestServerConfig()
	cfg.Mode, cfg.DataDir = consts.Development, t.TempDir()
	site, err := sitedb.NewDB(t.Context(), "qaaaaaaa", &cfg)
	require.NoError(t, err)
	defer site.Close()
	// The migrations go into the DB file, so the readings after are only in the write-ahead log
	require.NoError(t, site.Vacuum(t.Context()))
	for i := range 50 {
		_, err := site.Queries().CreateGlucoseReading(t.Context(), sitedb.CreateGlucoseReadingParams{ID: fmt.Sprint(i), Mgdl: 120, ReadingType: "sgv", OccurredAt: int64(i), CreatedByIdentityName: "phone"})
		require.NoError(t, err)
	}
	dir := t.TempDir()
	dbFileOnly, err := os.ReadFile(site.Path)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "db-file-only.sqlite"), dbFileOnly, 0o600))
	require.Zero(t, readings(t, filepath.Join(dir, "db-file-only.sqlite")))

	require.NoError(t, site.Backup(t.Context()))
	require.Equal(t, ".lz4", filepath.Ext(site.BackupPath))
	entries, err := os.ReadDir(filepath.Dir(site.BackupPath))
	require.NoError(t, err)
	require.Len(t, entries, 1)

	compressed, err := os.Open(site.BackupPath)
	require.NoError(t, err)
	defer compressed.Close()
	// Its frame is in 64 KB blocks: the block size is bits 4 to 6 of the frame descriptor's second byte, 4 for 64 KB
	frameStart := make([]byte, 6)
	_, err = io.ReadFull(compressed, frameStart)
	require.NoError(t, err)
	require.Equal(t, byte(4), frameStart[5]>>4&0x7)
	_, err = compressed.Seek(0, io.SeekStart)
	require.NoError(t, err)
	restored, err := os.Create(filepath.Join(dir, "restored.sqlite"))
	require.NoError(t, err)
	_, err = io.Copy(restored, lz4.NewReader(compressed))
	require.NoError(t, err)
	require.NoError(t, restored.Close())
	require.Equal(t, 50, readings(t, restored.Name()))
}

// A backup replaces the uncompressed copy kept before beside it
func TestCompressedBackupDeletesTheUncompressedCopy(t *testing.T) {
	cfg := *testutils.GetTestServerConfig()
	cfg.Mode, cfg.DataDir = consts.Development, t.TempDir()
	site, err := sitedb.NewDB(t.Context(), "qaaaaaaa", &cfg)
	require.NoError(t, err)
	defer site.Close()
	uncompressed := filepath.Join(filepath.Dir(site.BackupPath), "qaaaaaaa.sqlite")
	require.NoError(t, os.MkdirAll(filepath.Dir(uncompressed), 0o755))
	require.NoError(t, os.WriteFile(uncompressed, []byte("a backup from before"), 0o600))

	require.NoError(t, site.Backup(t.Context()))
	require.FileExists(t, site.BackupPath)
	require.NoFileExists(t, uncompressed)
}
