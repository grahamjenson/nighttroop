package backup

import (
	"context"
	"database/sql"
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

	err = conn.Raw(func(driverConn any) error {
		bck, err := driverConn.(BackupConn).NewBackup(backupPath)
		if err != nil {
			return err
		}
		for more := true; more; {
			more, err = bck.Step(-1)
			if err != nil {
				return err
			}
		}

		return bck.Finish()
	})

	if err != nil {
		slog.ErrorContext(ctx, "SqliteBackup Error on Raw", "backupPath", backupPath, "error", err)
		return fmt.Errorf("cannot backup db: %w", err)
	}

	return nil
}
