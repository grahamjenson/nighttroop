package errors

import (
	"errors"
	"strings"

	"modernc.org/sqlite"
	sqlite3 "modernc.org/sqlite/lib"
)

func IsSQLDuplicateID(err error) bool {
	// Exmaple Error String is "UNIQUE constraint failed: glucose_readings.id"
	// TODO: probably slow
	return IsSQLUniqueViolationError(err) && strings.Contains(err.Error(), ".id")
}

func IsSQLUniqueViolationError(err error) bool {
	var sqliteErr *sqlite.Error
	if errors.As(err, &sqliteErr) {
		if sqliteErr.Code() == sqlite3.SQLITE_CONSTRAINT_UNIQUE || sqliteErr.Code() == sqlite3.SQLITE_CONSTRAINT_PRIMARYKEY {
			return true
		}
	}
	return false
}

func IsSQLConstraintNotNull(err error) bool {
	var sqliteErr *sqlite.Error
	if !errors.As(err, &sqliteErr) {
		return false
	}

	return sqliteErr.Code() == sqlite3.SQLITE_CONSTRAINT_NOTNULL
}

func IsSQLConstraintCheck(err error) bool {
	var sqliteErr *sqlite.Error
	if !errors.As(err, &sqliteErr) {
		return false
	}

	return sqliteErr.Code() == sqlite3.SQLITE_CONSTRAINT_CHECK
}

func IsSQLConstraintForeignKey(err error) bool {
	var sqliteErr *sqlite.Error
	if !errors.As(err, &sqliteErr) {
		return false
	}

	return sqliteErr.Code() == sqlite3.SQLITE_CONSTRAINT_FOREIGNKEY
}

func IsSQLConstraintPrimaryKey(err error) bool {
	var sqliteErr *sqlite.Error
	if !errors.As(err, &sqliteErr) {
		return false
	}

	return sqliteErr.Code() == sqlite3.SQLITE_CONSTRAINT_PRIMARYKEY
}
