// try_migrate_or_delete opens each database in a data folder as the server does, migrating it and preparing its
// queries, so a dev server can keep its data between deploys. If any cannot be opened, as when a migration was edited
// rather than added, it deletes all of them, and their backups, so the server starts empty rather than with only some.
// The live DB is only ever in memory, so it has none.
//
//	try_migrate_or_delete /data
package main

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/data/sql/jobsdb"
	"nighttroop.t1d.nz/data/sql/logsdb"
	"nighttroop.t1d.nz/data/sql/sitedb"
)

// opener opens a kind of DB at a path as the server does, migrating it and preparing its queries
type opener func(ctx context.Context, path string) error

func open[Q interface{ Close() error }](openDB func(string, int) (*sql.DB, error), migrate func(*sql.DB) error, prepare func(context.Context, *sql.DB) (Q, error)) opener {
	return func(ctx context.Context, path string) error {
		db, err := openDB(path, 0)
		if err != nil {
			return err
		}
		defer db.Close()
		if err := migrate(db); err != nil {
			return err
		}
		queries, err := prepare(ctx, db)
		if err != nil {
			return err
		}
		return queries.Close()
	}
}

// openers are the DBs kept on disk by name
var openers = map[string]opener{
	"controldb": open(controldb.OpenSqliteDB, controldb.MigrateUp, func(ctx context.Context, db *sql.DB) (*controldb.Queries, error) { return controldb.Prepare(ctx, db) }),
	"jobsdb":    open(jobsdb.OpenSqliteDB, jobsdb.MigrateUp, func(ctx context.Context, db *sql.DB) (*jobsdb.Queries, error) { return jobsdb.Prepare(ctx, db) }),
	"logsdb":    open(logsdb.OpenSqliteDB, logsdb.MigrateUp, func(ctx context.Context, db *sql.DB) (*logsdb.Queries, error) { return logsdb.Prepare(ctx, db) }),
	"sitedb":    open(sitedb.OpenSqliteDB, sitedb.MigrateUp, func(ctx context.Context, db *sql.DB) (*sitedb.Queries, error) { return sitedb.Prepare(ctx, db) }),
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: try_migrate_or_delete <data folder>")
		os.Exit(2)
	}
	if err := tryMigrateOrDelete(context.Background(), os.Args[1]); err != nil {
		fmt.Fprintln(os.Stderr, "try_migrate_or_delete:", err)
		os.Exit(1)
	}
}

// tryMigrateOrDelete opens every DB in dataDir, deleting them all, and their backups, if one cannot be opened
func tryMigrateOrDelete(ctx context.Context, dataDir string) error {
	err := openAll(ctx, filepath.Join(dataDir, "db"))
	if err == nil {
		return nil
	}
	fmt.Println("try_migrate_or_delete: deleting every database, as one cannot be migrated:", err)
	for _, dir := range []string{"db", "backup"} {
		if err := os.RemoveAll(filepath.Join(dataDir, dir)); err != nil {
			return err
		}
	}
	return nil
}

// kindOf is the kind of DB a file in the DB folder is: a site's under site/, or the one it is named for
func kindOf(rel string) (string, bool) {
	if strings.HasPrefix(rel, "site/") {
		return "sitedb", true
	}
	name := strings.TrimSuffix(rel, ".sqlite")
	_, ok := openers[name]
	return name, ok
}

// openAll opens each DB in dir, which may not exist yet, stopping at the first that cannot be
func openAll(ctx context.Context, dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil
	}
	return filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".sqlite") {
			return err
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		name, ok := kindOf(filepath.ToSlash(rel))
		if !ok {
			return nil
		}
		if err := openers[name](ctx, path); err != nil {
			return fmt.Errorf("%s: %w", rel, err)
		}
		return nil
	})
}
