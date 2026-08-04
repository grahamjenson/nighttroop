package utils

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/mattn/go-sqlite3"
)

func SqliteBackup(ctx context.Context, db *sql.DB, backupPath string) error {
	defer TrackTime(time.Now(), "SqliteBackup "+backupPath, nil, nil)
	slog.Info("Starting SqliteBackup", "backupPath", backupPath)
	err := MkParentDirs(backupPath)
	if err != nil {
		return fmt.Errorf("cannot create parent directories: %w", err)
	}

	backupDB, err := sql.Open("sqlite3", backupPath)
	if err != nil {
		return fmt.Errorf("cannot open backup db: %w", err)
	}
	defer backupDB.Close()

	backupCon, err := backupDB.Conn(ctx)
	if err != nil {
		return fmt.Errorf("cannot get backup connection: %w", err)
	}
	defer backupCon.Close()

	srcConn, err := db.Conn(ctx)
	if err != nil {
		return fmt.Errorf("cannot get source connection: %w", err)
	}
	defer srcConn.Close()

	err = backupCon.Raw(func(backupConRaw any) error {
		backupConSQLite, ok := backupConRaw.(*sqlite3.SQLiteConn)
		if !ok {
			slog.Error("Error getting SQLiteConn for backupConSQLite", "backupPath", backupPath, "error", err)

			return fmt.Errorf("failed to get SQLiteConn backupConSQLite")
		}
		srcConn.Raw(func(srcConnRaw any) error {

			srcConnSQLite, ok := srcConnRaw.(*sqlite3.SQLiteConn)
			if !ok {
				slog.Error("Error getting SQLiteConn for srcConnSQLite", "backupPath", backupPath, "error", err)

				return fmt.Errorf("failed to get SQLiteConn srcConnSQLite")
			}

			backup, err := backupConSQLite.Backup("main", srcConnSQLite, "main")
			if err != nil {
				slog.Error("Error creating backup", "backupPath", backupPath, "error", err)
				return fmt.Errorf("cannot create backup no create: %w", err)
			}

			defer backup.Finish()
			_, err = backup.Step(-1)

			if err != nil {
				slog.Error("Error stepping backup", "backupPath", backupPath, "error", err)
				return fmt.Errorf("cannot create backup no step 1: %w", err)
			}

			return nil
		})
		return nil
	})
	if err != nil {
		slog.Error("SqliteBackup Error on Raw", "backupPath", backupPath, "error", err)
		return fmt.Errorf("cannot backup db: %w", err)
	}

	return nil
}
