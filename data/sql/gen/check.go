//go:build ignore

package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"

	_ "modernc.org/sqlite"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/data/sql/gen"
	"nighttroop.t1d.nz/data/sql/jobsdb"
	"nighttroop.t1d.nz/data/sql/livedb"
	"nighttroop.t1d.nz/data/sql/logsdb"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/consts"
	utilerrors "nighttroop.t1d.nz/utils/errors"
)

const topwrapper string = "\n\n------------------------------ SQL CHECK ERRORS -----------------------------------------------------------\n\n"
const bottomwrapper string = "\n\n-----------------------------------------------------------------------------------------------------------\n\n"

func main() {
	ctx := context.Background()
	config := config.ServerConfig{
		Mode: consts.Test, // Get In memory Database
	}

	siteDB, err := sitedb.NewDB(ctx, "subdomain", &config)
	if err != nil {
		log.Fatalf("%sfailed to create site DB: %v %s", topwrapper, err, bottomwrapper)
	}

	controlDB, err := controldb.NewDB(ctx, &config)
	if err != nil {
		log.Fatalf("%sfailed to create control DB: %v %s", topwrapper, err, bottomwrapper)
	}

	jobsDB, err := jobsdb.NewDB(ctx, &config)
	if err != nil {
		log.Fatalf("%sfailed to create jobs DB: %v %s", topwrapper, err, bottomwrapper)
	}

	logsDB, err := logsdb.NewDB(ctx, &config)
	if err != nil {
		log.Fatalf("%sfailed to create logs DB: %v %s", topwrapper, err, bottomwrapper)
	}

	liveDB, err := livedb.NewDB(ctx, &config)
	if err != nil {
		log.Fatalf("%sfailed to create live DB: %v %s", topwrapper, err, bottomwrapper)
	}

	/////////
	// EXECUTE ALL QUERIES
	/////////

	// Sometimes SQLc will generate a query that is incorrect causing a syntax error when queried.
	// This executes all queries ignoring normal errors and checking for a bad query syntax.
	// A site DB runs its queries unprepared, so they are prepared here to fail on any that cannot be
	siteQueries, err := sitedb.Prepare(ctx, siteDB.SQLiteDB)
	if err != nil {
		log.Fatalf("%sfailed to prepare site DB queries: %v %s", topwrapper, err, bottomwrapper)
	}
	if err := executeAllQueries(ctx, siteQueries); err != nil {
		log.Fatalf("%sfailed to create site DB: %v %s", topwrapper, err, bottomwrapper)
	}

	if err := executeAllQueries(ctx, controlDB.Queries); err != nil {
		log.Fatalf("%sfailed to create control DB: %v %s", topwrapper, err, bottomwrapper)
	}
	if err := executeAllQueries(ctx, jobsDB.Queries); err != nil {
		log.Fatalf("%sfailed to create jobs DB: %v %s", topwrapper, err, bottomwrapper)
	}
	if err := executeAllQueries(ctx, logsDB.Queries); err != nil {
		log.Fatalf("%sfailed to create logs DB: %v %s", topwrapper, err, bottomwrapper)
	}
	if err := executeAllQueries(ctx, liveDB.Queries); err != nil {
		log.Fatalf("%sfailed to create live DB: %v %s", topwrapper, err, bottomwrapper)
	}

	////////
	// CHECK ALL TABLES ARE STRICT
	////////
	for _, dbi := range []interface {
		Name() string
		GetRawDB() *sql.DB
	}{siteDB, controlDB, jobsDB, logsDB, liveDB} {
		if err := assertAllTablesStrict(dbi.GetRawDB()); err != nil {
			log.Fatalf("%s%s DB: %v %s", topwrapper, dbi.Name(), err, bottomwrapper)
		}
		if err := assertJSONColumnsValid(dbi.GetRawDB()); err != nil {
			log.Fatalf("%s%s DB: %v %s", topwrapper, dbi.Name(), err, bottomwrapper)
		}
	}

	////////
	/// CHECK MIGRATIONS DOWN
	/////////
	if err := checkMigrateDownDB(siteDB); err != nil {
		log.Fatalf("%scheck migrate down site DB failed: %v %s", topwrapper, err, bottomwrapper)
	}

	if err := checkMigrateDownDB(controlDB); err != nil {
		log.Fatalf("%scheck migrate down control DB failed: %v %s", topwrapper, err, bottomwrapper)
	}

	if err := checkMigrateDownDB(jobsDB); err != nil {
		log.Fatalf("%scheck migrate down jobs DB failed: %v %s", topwrapper, err, bottomwrapper)
	}

	if err := checkMigrateDownDB(logsDB); err != nil {
		log.Fatalf("%scheck migrate down logs DB failed: %v %s", topwrapper, err, bottomwrapper)
	}

	if err := checkMigrateDownDB(liveDB); err != nil {
		log.Fatalf("%scheck migrate down live DB failed: %v %s", topwrapper, err, bottomwrapper)
	}

	////////
	/// CHECK MIGRATIONS WITH ROWS AT THEIR LIMITS
	////////
	// Each DB on disk, opened as the server opens it, with foreign keys on
	dir, err := os.MkdirTemp("", "check-migrations-")
	if err != nil {
		log.Fatalf("%scannot make a folder to check migrations in: %v %s", topwrapper, err, bottomwrapper)
	}
	defer os.RemoveAll(dir)
	for _, d := range []struct {
		name       string
		migrations fs.FS
		open       func(string, int) (*sql.DB, error)
	}{
		{"site", sitedb.Migrations, sitedb.OpenSqliteDB},
		{"control", controldb.Migrations, controldb.OpenSqliteDB},
		{"jobs", jobsdb.Migrations, jobsdb.OpenSqliteDB},
		{"logs", logsdb.Migrations, logsdb.OpenSqliteDB},
		{"live", livedb.Migrations, livedb.OpenSqliteDB},
	} {
		db, err := d.open(filepath.Join(dir, d.name+".sqlite"), 0)
		if err == nil {
			err = gen.CheckMigrationsWithData(d.migrations, db)
		}
		if err != nil {
			log.Fatalf("%scheck migrations of the %s DB with rows at their limits failed: %v %s", topwrapper, d.name, err, bottomwrapper)
		}
	}
}

func checkMigrateDownDB(dbi interface {
	MigrateDown() error
	Name() string
	GetRawDB() *sql.DB
}) error {

	err := dbi.MigrateDown()
	if err != nil {
		return fmt.Errorf("failed to migrate down %s db: %w", dbi.Name(), err)
	}

	return nil
}

// assertAllTablesStrict fails if any table created by the migrations is not
// declared STRICT, so column types are enforced instead of being affinities.
func assertAllTablesStrict(db *sql.DB) error {
	rows, err := db.Query(`
		SELECT name, sql
		FROM sqlite_master
		WHERE type = 'table'
		  AND name NOT LIKE 'sqlite_%'
		  AND name <> 'schema_migrations'
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	var notStrict []string
	for rows.Next() {
		var name, createSQL string
		if err := rows.Scan(&name, &createSQL); err != nil {
			return err
		}
		// Table options follow the closing parenthesis of the column list.
		options := createSQL[strings.LastIndex(createSQL, ")")+1:]
		if !strings.Contains(strings.ToUpper(options), "STRICT") {
			notStrict = append(notStrict, name)
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}

	if len(notStrict) > 0 {
		return fmt.Errorf("tables must be created STRICT: %v", notStrict)
	}
	return nil
}

// assertJSONColumnsValid fails if a column holding JSON (named with json, a list of errors, or
// attributes) does not CHECK json_valid on its own line, so invalid JSON is refused when written
// rather than failing when read.
func assertJSONColumnsValid(db *sql.DB) error {
	rows, err := db.Query(`
		SELECT name, sql
		FROM sqlite_master
		WHERE type = 'table'
		  AND name NOT LIKE 'sqlite_%'
		  AND name <> 'schema_migrations'
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	var missing []string
	for rows.Next() {
		var name, createSQL string
		if err := rows.Scan(&name, &createSQL); err != nil {
			return err
		}
		for _, line := range strings.Split(createSQL, "\n") {
			fields := strings.Fields(line)
			if len(fields) < 2 || !strings.EqualFold(fields[1], "TEXT") {
				continue
			}
			column := fields[0]
			isJSON := strings.Contains(column, "json") || strings.HasSuffix(column, "errors") || column == "attributes"
			if isJSON && !strings.Contains(line, "json_valid("+column+")") {
				missing = append(missing, name+"."+column)
			}
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}

	if len(missing) > 0 {
		return fmt.Errorf("JSON columns must CHECK json_valid: %v", missing)
	}
	return nil
}

func assertNoTables(db *sql.DB) error {
	rows, err := db.Query(`
		SELECT name
		FROM sqlite_master
		WHERE type IN ('table', 'index', 'trigger', 'view')
		  AND name NOT LIKE 'sqlite_%'
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	var objects []string

	skip := []string{"schema_migrations", "version_unique"}
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return err
		}

		if slices.Contains(skip, name) {
			continue
		}
		objects = append(objects, name)
	}

	if err := rows.Err(); err != nil {
		return err
	}

	if len(objects) > 0 {
		return fmt.Errorf("database not empty after migrate down: %v", objects)
	}

	return nil
}

var ignoredQueries = map[string]struct{}{
	"Close":  {},
	"WithTx": {},
}

func executeAllQueries(ctx context.Context, queries any) error {
	v := reflect.ValueOf(queries)

	for i := 0; i < v.NumMethod(); i++ {
		method := v.Type().Method(i)
		fn := v.Method(i)

		if _, ignored := ignoredQueries[method.Name]; ignored {
			continue
		}

		if err := executeQuery(ctx, fn); err != nil {
			return fmt.Errorf("%s: %w", method.Name, err)
		}
	}

	return nil
}

func executeQuery(ctx context.Context, fn reflect.Value) error {
	fnType := fn.Type()

	// Every query should be:
	//
	// func(context.Context, SomeParams) (..., error)
	if fnType.NumIn() != 1 && fnType.NumIn() != 2 {
		return fmt.Errorf("expected 1 or 2 arguments, got %d", fnType.NumIn())
	}

	args := []reflect.Value{
		reflect.ValueOf(ctx),
	}
	if fnType.NumIn() == 2 {
		args = append(args, reflect.Zero(fnType.In(1)))
	}

	results := fn.Call(args)

	// Last return value should be error.
	if len(results) == 0 {
		return fmt.Errorf("query returns no values")
	}

	last := results[len(results)-1]

	if last.Type() != reflect.TypeOf((*error)(nil)).Elem() {
		return fmt.Errorf("last return value is %v, not error", last.Type())
	}

	if !last.IsNil() {
		err := last.Interface().(error)

		// No rows is a valid result for SELECT queries.
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}

		if utilerrors.IsSQLDuplicateID(err) {
			return nil
		}
		if utilerrors.IsSQLUniqueViolationError(err) {
			return nil
		}
		if utilerrors.IsSQLConstraintNotNull(err) {
			return nil
		}
		if utilerrors.IsSQLConstraintCheck(err) {
			return nil
		}
		if utilerrors.IsSQLConstraintForeignKey(err) {
			return nil
		}
		if utilerrors.IsSQLConstraintPrimaryKey(err) {
			return nil
		}
		return err
	}

	return nil
}

func isQueryMethod(t reflect.Type) bool {
	contextType := reflect.TypeOf((*context.Context)(nil)).Elem()

	if t.NumIn() < 1 || t.NumIn() > 2 {
		return false
	}

	// First argument must always be context.Context.
	if t.In(0) != contextType {
		return false
	}

	// Optional second argument must be a params struct.
	if t.NumIn() == 2 && t.In(1).Kind() != reflect.Struct {
		return false
	}

	return true
}
