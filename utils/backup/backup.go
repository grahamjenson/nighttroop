package backup

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"log/slog"

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
		// A connection wrapped, e.g. to time its queries, backs up as the driver's own
		if wrapped, ok := driverConn.(interface{ Unwrap() driver.Conn }); ok {
			driverConn = wrapped.Unwrap()
		}
		backupConn, ok := driverConn.(BackupConn)
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
