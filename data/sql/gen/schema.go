package gen

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

func BuildSchema(outputSchema string, migrationFiles []string) error {

	tmpFile, err := os.CreateTemp("", "sqlc-schema-*.db")
	if err != nil {
		return fmt.Errorf("failed to create temp db: %w", err)
	}

	tmpDB := tmpFile.Name()
	defer os.Remove(tmpDB)

	// Remove existing temp DB
	if err := os.Remove(tmpDB); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove temp db: %w", err)
	}

	// Open SQLite database
	db, err := sql.Open("sqlite3", tmpDB)
	if err != nil {
		return fmt.Errorf("failed to open sqlite db: %w", err)
	}
	defer db.Close()

	// Apply migrations
	for _, file := range migrationFiles {
		fmt.Printf("Applying %s\n", file)

		sqlBytes, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to read migration %s: %w", file, err)
		}

		if _, err := db.Exec(string(sqlBytes)); err != nil {
			return fmt.Errorf("failed to execute migration %s: %w", file, err)
		}
	}

	// Read schema from sqlite_master
	rows, err := db.Query(`
		SELECT sql
		FROM sqlite_master
		WHERE sql IS NOT NULL
		  AND type IN ('table', 'index', 'trigger', 'view')
		ORDER BY type, name;
	`)
	if err != nil {
		return fmt.Errorf("failed to query schema: %w", err)
	}
	defer rows.Close()

	// Create schema output file
	out, err := os.Create(outputSchema)
	if err != nil {
		return fmt.Errorf("failed to create schema file: %w", err)
	}
	defer out.Close()

	// Write schema statements
	for rows.Next() {
		var stmt string

		if err := rows.Scan(&stmt); err != nil {
			return fmt.Errorf("failed to scan schema row: %w", err)
		}

		if _, err := fmt.Fprintf(out, "%s;\n\n", stmt); err != nil {
			return fmt.Errorf("failed to write schema: %w", err)
		}
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("schema iteration error: %w", err)
	}

	fmt.Printf("Generated %s\n", outputSchema)

	return nil
}
