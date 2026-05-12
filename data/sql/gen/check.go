//go:build ignore

package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
	"nighttroop.t1d.nz/data/sql/control"
	controlMigrations "nighttroop.t1d.nz/data/sql/control/migrations"
	"nighttroop.t1d.nz/data/sql/user"
	userMigrations "nighttroop.t1d.nz/data/sql/user/migrations"
)

func main() {
	if err := checkUsers(); err != nil {
		log.Fatalf("check users failed: %v", err)
	}

	if err := checkControl(); err != nil {
		log.Fatalf("check control failed: %v", err)
	}

}

func checkUsers() error {
	usersTmpFile, err := os.CreateTemp("", "sqlc-schema-*.db")
	if err != nil {
		return fmt.Errorf("failed to create temp db: %w", err)
	}

	usersTmpFileDB := usersTmpFile.Name()
	defer os.Remove(usersTmpFileDB)

	// Remove existing temp DB
	if err := os.Remove(usersTmpFileDB); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove temp db: %w", err)
	}

	// Open SQLite database
	db, err := sql.Open("sqlite3", usersTmpFileDB)
	if err != nil {
		return fmt.Errorf("failed to open sqlite db: %w", err)
	}
	defer db.Close()

	err = userMigrations.MigrateUser(context.TODO(), db)
	if err != nil {
		return fmt.Errorf("failed to migrate user db: %w", err)
	}

	_, err = user.Prepare(context.TODO(), db)
	if err != nil {
		return fmt.Errorf("failed to prepare user queries: %w", err)
	}

	return nil
}

func checkControl() error {
	usersTmpFile, err := os.CreateTemp("", "sqlc-schema-*.db")
	if err != nil {
		return fmt.Errorf("failed to create temp db: %w", err)
	}

	usersTmpFileDB := usersTmpFile.Name()
	defer os.Remove(usersTmpFileDB)

	// Remove existing temp DB
	if err := os.Remove(usersTmpFileDB); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove temp db: %w", err)
	}

	// Open SQLite database
	db, err := sql.Open("sqlite3", usersTmpFileDB)
	if err != nil {
		return fmt.Errorf("failed to open sqlite db: %w", err)
	}
	defer db.Close()

	err = controlMigrations.MigrateControl(context.TODO(), db)
	if err != nil {
		return fmt.Errorf("failed to migrate user db: %w", err)
	}

	_, err = control.Prepare(context.TODO(), db)
	if err != nil {
		return fmt.Errorf("failed to prepare control queries: %w", err)
	}

	return nil
}
