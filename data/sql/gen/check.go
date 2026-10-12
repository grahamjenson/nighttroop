//go:build ignore

package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"reflect"
	"slices"

	_ "modernc.org/sqlite"
	"nighttroop.t1d.nz/data/sql/control"
	"nighttroop.t1d.nz/data/sql/jobs"
	"nighttroop.t1d.nz/data/sql/site"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/consts"
	utilerrors "nighttroop.t1d.nz/utils/errors"
)

const topwrapper string = "\n\n------------------------------ SQL CHECK ERRORS -----------------------------------------------------------\n\n"
const bottomwrapper string = "\n\n-----------------------------------------------------------------------------------------------------------\n\n"

func main() {
	config := config.ServerConfig{
		Mode: consts.Test, // Get In memory Database
	}

	siteDB, err := site.NewDB(context.TODO(), "subdomain", &config)
	if err != nil {
		log.Fatalf("%sfailed to create site DB: %v %s", topwrapper, err, bottomwrapper)
	}

	controlDB, err := control.NewDB(context.TODO(), &config)
	if err != nil {
		log.Fatalf("%sfailed to create control DB: %v %s", topwrapper, err, bottomwrapper)
	}

	jobsDB, err := jobs.NewDB(context.TODO(), &config)
	if err != nil {
		log.Fatalf("%sfailed to create jobs DB: %v %s", topwrapper, err, bottomwrapper)
	}

	/////////
	// EXECUTE ALL QUERIES
	/////////

	// Sometimes SQLc will generate a query that is incorrect causing a syntax error when queried.
	// This executes all queries ignoring normal errors and checking for a bad query syntax.
	if err := executeAllQueries(context.TODO(), siteDB.Queries); err != nil {
		log.Fatalf("%sfailed to create site DB: %v %s", topwrapper, err, bottomwrapper)
	}

	if err := executeAllQueries(context.TODO(), controlDB.Queries); err != nil {
		log.Fatalf("%sfailed to create control DB: %v %s", topwrapper, err, bottomwrapper)
	}
	if err := executeAllQueries(context.TODO(), jobsDB.Queries); err != nil {
		log.Fatalf("%sfailed to create jobs DB: %v %s", topwrapper, err, bottomwrapper)
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
