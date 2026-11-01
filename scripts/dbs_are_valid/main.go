// dbs_are_valid says whether the databases in a data folder still fit the server's migrations, so a dev server can keep
// its data between deploys and wipe it only when it cannot be used. Each DB is migrated as the server would migrate
// it, then its schema is compared with the schema its migrations make for a new DB: an edited migration, which keeps
// its version, makes them differ. It exits 0 when every DB fits, or there are none, and 1 when one does not.
//
//	dbs_are_valid /data
package main

import (
	"context"
	"database/sql"
	"errors"
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

// kind is how one kind of DB is opened, migrated and read
type kind struct {
	open    func(path string, cacheSize int) (*sql.DB, error)
	memory  func() (*sql.DB, error)
	migrate func(*sql.DB) error
	schema  func(context.Context, *sql.DB) ([]string, error)
}

// kinds are the DBs kept on disk by name; the live DB is only ever in memory
var kinds = map[string]kind{
	"controldb": {controldb.OpenSqliteDB, controldb.OpenInMemorySqliteDB, controldb.MigrateUp, controldb.Schema},
	"jobsdb":    {jobsdb.OpenSqliteDB, jobsdb.OpenInMemorySqliteDB, jobsdb.MigrateUp, jobsdb.Schema},
	"logsdb":    {logsdb.OpenSqliteDB, logsdb.OpenInMemorySqliteDB, logsdb.MigrateUp, logsdb.Schema},
	"sitedb":    {sitedb.OpenSqliteDB, sitedb.OpenInMemorySqliteDB, sitedb.MigrateUp, sitedb.Schema},
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: dbs_are_valid <data folder>")
		os.Exit(2)
	}
	if err := check(context.Background(), filepath.Join(os.Args[1], "db")); err != nil {
		fmt.Fprintln(os.Stderr, "dbs_are_valid:", err)
		os.Exit(1)
	}
}

// kindOf is the kind of DB a file in the DB folder is: a site's under site/, or the one it is named for
func kindOf(rel string) (string, bool) {
	if strings.HasPrefix(rel, "site/") {
		return "sitedb", true
	}
	name := strings.TrimSuffix(rel, ".sqlite")
	_, ok := kinds[name]
	return name, ok
}

// check is nil when every DB in dir, which may not exist yet, fits its migrations
func check(ctx context.Context, dir string) error {
	want := map[string][]string{}
	return filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if path == dir && errors.Is(err, fs.ErrNotExist) {
			return nil
		}
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
		k := kinds[name]
		if want[name] == nil {
			if want[name], err = schemaOf(ctx, k, k.memory); err != nil {
				return fmt.Errorf("%s's migrations: %w", name, err)
			}
		}
		got, err := schemaOf(ctx, k, func() (*sql.DB, error) { return k.open(path, 0) })
		if err != nil {
			return fmt.Errorf("%s cannot be migrated: %w", rel, err)
		}
		if !slices.Equal(got, want[name]) {
			return fmt.Errorf("%s does not fit its migrations", rel)
		}
		return nil
	})
}

// schemaOf is the schema of the DB open opens, once migrated
func schemaOf(ctx context.Context, k kind, open func() (*sql.DB, error)) ([]string, error) {
	db, err := open()
	if err != nil {
		return nil, err
	}
	defer db.Close()
	if err := k.migrate(db); err != nil {
		return nil, err
	}
	return k.schema(ctx, db)
}
