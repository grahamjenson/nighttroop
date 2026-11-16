package backup

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/pierrec/lz4/v4"
	"modernc.org/sqlite"
	"nighttroop.t1d.nz/utils"
)

type BackupConn interface {
	NewBackup(string) (*sqlite.Backup, error)
}

// SqliteBackup copies db to backupPath. It logs only when it fails, as every site's DB is backed up in turn.
func SqliteBackup(ctx context.Context, db *sql.DB, backupPath string) error {
	err := utils.MkParentDirs(backupPath)
	if err != nil {
		return fmt.Errorf("cannot create parent directories: %w", err)
	}

	conn, err := db.Conn(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "SqliteBackup Error on Conn", "backupPath", backupPath, "error", err)
		return err
	}
	defer conn.Close()

	err = conn.Raw(func(driverConn any) (err error) {
		backupConn, ok := unwrap(driverConn).(BackupConn)
		if !ok {
			return fmt.Errorf("the db's connection, a %T, cannot back up", driverConn)
		}
		bck, err := backupConn.NewBackup(backupPath)
		if err != nil {
			return err
		}
		// Finished even when a step fails, or the destination stays open and locked, failing every later backup to it
		defer func() {
			if finishErr := bck.Finish(); err == nil {
				err = finishErr
			}
		}()
		for more := true; more && err == nil; {
			more, err = bck.Step(-1)
		}
		return err
	})

	if err != nil {
		slog.ErrorContext(ctx, "SqliteBackup Error on Raw", "backupPath", backupPath, "error", err)
		return fmt.Errorf("cannot backup db: %w", err)
	}

	return nil
}

// unwrap is the driver's own connection of one wrapped, e.g. to time its queries, so it backs up as the driver's own
func unwrap(driverConn any) any {
	if wrapped, ok := driverConn.(interface{ Unwrap() driver.Conn }); ok {
		return wrapped.Unwrap()
	}
	return driverConn
}

type serializer interface {
	Serialize() ([]byte, error)
}

// CompressedBackup writes a consistent copy of db, with what is only in its write-ahead log, to backupPath, compressed
// with LZ4. The copy is in memory while it is compressed, so it is for DBs of tens of MB, as a site's. It is written to
// a temp file beside backupPath, then renamed, so a copy cut short is never at backupPath. It logs only when it fails,
// as every site's DB is backed up in turn.
func CompressedBackup(ctx context.Context, db *sql.DB, backupPath string) error {
	if err := utils.MkParentDirs(backupPath); err != nil {
		return fmt.Errorf("cannot create parent directories: %w", err)
	}
	conn, err := db.Conn(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	var image []byte
	err = conn.Raw(func(driverConn any) (err error) {
		s, ok := unwrap(driverConn).(serializer)
		if !ok {
			return fmt.Errorf("the db's connection, a %T, cannot be copied", driverConn)
		}
		image, err = s.Serialize()
		// SQLite gives no copy, and the driver no error, when it cannot allocate one
		if err == nil && len(image) == 0 {
			err = errors.New("the db's copy is empty")
		}
		return err
	})
	if err == nil {
		err = writeCompressed(image, backupPath)
	}
	if err != nil {
		slog.ErrorContext(ctx, "CompressedBackup failed", "backupPath", backupPath, "error", err)
		return fmt.Errorf("cannot backup db: %w", err)
	}
	return nil
}

// writeCompressed writes data compressed with LZ4 to a temp file beside path, then renames it to path
func writeCompressed(data []byte, path string) error {
	file, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+"-*")
	if err != nil {
		return err
	}
	// Once renamed there is nothing to remove
	defer os.Remove(file.Name())
	lz := lz4.NewWriter(file)
	// 64 KB blocks, not LZ4's default 4 MB: each backup held 4 MB buffers, 16% of the live heap under load on staging
	err = lz.Apply(lz4.BlockSizeOption(lz4.Block64Kb))
	if err == nil {
		_, err = lz.Write(data)
	}
	if err == nil {
		err = lz.Close()
	}
	if closeErr := file.Close(); err == nil {
		err = closeErr
	}
	if err == nil {
		err = os.Rename(file.Name(), path)
	}
	return err
}
