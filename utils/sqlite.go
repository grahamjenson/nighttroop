package utils

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"modernc.org/sqlite"
)

type BackupConn interface {
	NewBackup(string) (*sqlite.Backup, error)
}

func SqliteBackup(ctx context.Context, db *sql.DB, backupPath string) error {
	defer TrackTime(time.Now(), "SqliteBackup "+backupPath, nil, nil)
	slog.Info("Starting SqliteBackup", "backupPath", backupPath)
	err := MkParentDirs(backupPath)
	if err != nil {
		return fmt.Errorf("cannot create parent directories: %w", err)
	}

	conn, err := db.Conn(ctx)
	if err != nil {
		slog.Error("SqliteBackup Error on Conn", "backupPath", backupPath, "error", err)
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
		slog.Error("SqliteBackup Error on Raw", "backupPath", backupPath, "error", err)
		return fmt.Errorf("cannot backup db: %w", err)
	}

	return nil
}
