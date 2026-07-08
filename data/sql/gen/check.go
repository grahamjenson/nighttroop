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

	"nighttroop.t1d.nz/data/sql/site"
)

func main() {
	if err := checkSites(); err != nil {
		log.Fatalf("check sites failed: %v", err)
	}

	if err := checkControl(); err != nil {
		log.Fatalf("check control failed: %v", err)
	}

}

func checkSites() error {
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

	err = site.MigrateUp("dummy", db)
	if err != nil {
		return fmt.Errorf("failed to migrate user db: %w", err)
	}

	_, err = site.Prepare(context.TODO(), db)
	if err != nil {
		return fmt.Errorf("failed to prepare user queries: %w", err)
	} else {
		fmt.Println("USER DB Prepared successfully")
	}

	err = site.MigrateDown("dummy", db)
	if err != nil {
		return fmt.Errorf("failed to migrate down user db: %w", err)
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

	////
	// MIGTATE UP THE DB
	////
	err = control.MigrateUp("dummy", db)
	if err != nil {
		return fmt.Errorf("failed to migrate control db: %w", err)
	}

	//// PREPARE THE QUERIES
	_, err = control.Prepare(context.TODO(), db)
	if err != nil {
		return fmt.Errorf("failed to prepare control queries: %w", err)
	} else {
		fmt.Println("Control DB Prepared successfully")
	}

	////
	// MIGRATE DOWN THE DB
	////
	err = control.MigrateDown("dummy", db)
	if err != nil {
		return fmt.Errorf("failed to migrate down control db: %w", err)
	}

	return nil
}
