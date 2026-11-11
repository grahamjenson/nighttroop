// try_migrate_or_delete opens each database in a data folder as the server does, migrating it, checking its schema
// and preparing its queries, once for the site databases, so a dev server can keep its data between deploys. If any
// cannot be opened, or its schema is not the one its migrations make on a new database, as when a migration was edited
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
	"slices"
	"strings"

	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/data/sql/jobsdb"
	"nighttroop.t1d.nz/data/sql/logsdb"
	"nighttroop.t1d.nz/data/sql/sitedb"
)

// opener opens a kind of DB at a path as the server does, migrating it, checking its schema and preparing its queries
type opener func(ctx context.Context, path string) error

// kind is how a kind of DB is opened on disk and in memory, migrated, and its schema read
type kind struct {
	openDB    func(string, int) (*sql.DB, error)
	openNewDB func() (*sql.DB, error)
	migrate   func(*sql.DB) error
	schema    func(context.Context, *sql.DB) (string, error)
	newSchema string // the schema its migrations make on a new DB, once read
}

// checkSchema fails when db's schema is not the one k's migrations make on a new DB in memory, as a DB made before a
// migration was edited keeps the old one even when every query still works on it, such as with a changed CHECK
func (k *kind) checkSchema(ctx context.Context, db *sql.DB) error {
	if k.newSchema == "" {
		newDB, err := k.openNewDB()
		if err != nil {
			return err
		}
		defer newDB.Close()
		if err := k.migrate(newDB); err != nil {
			return err
		}
		if k.newSchema, err = k.schema(ctx, newDB); err != nil {
			return err
		}
	}
	schema, err := k.schema(ctx, db)
	if err != nil {
		return err
	}
	if schema == k.newSchema {
		return nil
	}
	return fmt.Errorf("its schema is not the one its migrations make: %s", difference(schema, k.newSchema))
}

// difference is a statement in one schema and not the other, to say why they differ
func difference(schema, newSchema string) string {
	stmts, newStmts := strings.SplitAfter(schema, ";\n"), strings.SplitAfter(newSchema, ";\n")
	for _, stmt := range stmts {
		if !slices.Contains(newStmts, stmt) {
			return fmt.Sprintf("it has %q", stmt)
		}
	}
	for _, stmt := range newStmts {
		if !slices.Contains(stmts, stmt) {
			return fmt.Sprintf("it has no %q", stmt)
		}
	}
	return "its statements are in another order"
}

func open[Q interface{ Close() error }](k *kind, prepare func(context.Context, *sql.DB) (Q, error)) opener {
	return func(ctx context.Context, path string) error {
		db, err := k.openDB(path, 0)
		if err != nil {
			return err
		}
		defer db.Close()
		if err := k.migrate(db); err != nil {
			return err
		}
		if err := k.checkSchema(ctx, db); err != nil {
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
	"controldb": open(&kind{openDB: controldb.OpenSqliteDB, openNewDB: controldb.OpenInMemorySqliteDB, migrate: controldb.MigrateUp, schema: controldb.Schema},
		func(ctx context.Context, db *sql.DB) (*controldb.Queries, error) { return controldb.Prepare(ctx, db) }),
	"jobsdb": open(&kind{openDB: jobsdb.OpenSqliteDB, openNewDB: jobsdb.OpenInMemorySqliteDB, migrate: jobsdb.MigrateUp, schema: jobsdb.Schema},
		func(ctx context.Context, db *sql.DB) (*jobsdb.Queries, error) { return jobsdb.Prepare(ctx, db) }),
	"logsdb": open(&kind{openDB: logsdb.OpenSqliteDB, openNewDB: logsdb.OpenInMemorySqliteDB, migrate: logsdb.MigrateUp, schema: logsdb.Schema},
		func(ctx context.Context, db *sql.DB) (*logsdb.Queries, error) { return logsdb.Prepare(ctx, db) }),
	"sitedb": open(&kind{openDB: sitedb.OpenSqliteDB, openNewDB: sitedb.OpenInMemorySqliteDB, migrate: sitedb.MigrateUp, schema: sitedb.Schema},
		prepareSite),
}

// sitePrepared is whether a site DB's queries have been prepared. There can be thousands of sites, every one with the
// schema checkSchema checked, and preparing is most of the time opening one takes.
var sitePrepared bool

// prepareSite prepares a site DB's queries, unless another's have been
func prepareSite(ctx context.Context, db *sql.DB) (*sitedb.Queries, error) {
	if sitePrepared {
		return sitedb.New(db), nil
	}
	queries, err := sitedb.Prepare(ctx, db)
	if err != nil {
		return nil, err
	}
	sitePrepared = true
	return queries, nil
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
