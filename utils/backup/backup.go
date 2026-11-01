package backup

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"modernc.org/sqlite"
	"nighttroop.t1d.nz/utils"
	"nighttroop.t1d.nz/utils/aslog"
)

type BackupConn interface {
	NewBackup(string) (*sqlite.Backup, error)
}

func SqliteBackup(ctx context.Context, db *sql.DB, backupPath string) error {
	defer aslog.TrackTime(ctx, time.Now(), "SqliteBackup "+backupPath, nil, nil)
	slog.InfoContext(ctx, "Starting SqliteBackup", "backupPath", backupPath)
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
		bck, err := driverConn.(BackupConn).NewBackup(backupPath)
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
