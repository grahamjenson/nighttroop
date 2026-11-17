// Package migrateup migrates each DB up with golang-migrate when it is opened, first recovering one a failed migration
// left dirty
package migrateup

import (
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	"github.com/golang-migrate/migrate/v4/source"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

// Up migrates db, the DB called name, up with the migrations in fsys's migrations folder. golang-migrate marks a DB
// dirty at a version before running its migration, and refuses a dirty DB after. Each migration runs in one
// transaction, so a DB dirty at version N had none of N applied: it is forced back to the version before N and migrated
// again. A migration that fails gives an error naming the DB, the migration and why it failed.
func Up(name string, db *sql.DB, fsys fs.FS) error {
	src, err := iofs.New(fsys, "migrations")
	if err != nil {
		return err
	}
	driver, err := sqlite.WithInstance(db, &sqlite.Config{})
	if err != nil {
		return err
	}
	m, err := migrate.NewWithInstance("iofs", src, "sqlite", driver)
	if err != nil {
		return err
	}
	if err := forceBackIfDirty(name, m, src); err != nil {
		return err
	}

	err = m.Up()
	if err == nil || errors.Is(err, migrate.ErrNoChange) {
		return nil
	}
	// A failed migration's error holds the whole file it ran, so only why it failed is kept
	var dbErr *database.Error
	if errors.As(err, &dbErr) && dbErr.OrigErr != nil {
		err = dbErr.OrigErr
	}
	// Dirty at the migration that failed, or not dirty at one this release does not have, as a DB from a later release
	if version, _, verr := m.Version(); verr == nil {
		return &FailedError{DB: name, Migration: version, Err: err}
	}
	return fmt.Errorf("%s cannot migrate: %w", name, err)
}

// forceBackIfDirty forces a DB dirty at a version back to the version before it, or to none when it is the first
func forceBackIfDirty(name string, m *migrate.Migrate, src source.Driver) error {
	version, dirty, err := m.Version()
	if errors.Is(err, migrate.ErrNilVersion) || (err == nil && !dirty) {
		return nil
	}
	if err != nil {
		return err
	}

	prev := database.NilVersion
	if first, err := src.First(); err != nil || first != version {
		p, err := src.Prev(version)
		if err != nil {
			return &FailedError{DB: name, Migration: version, Err: fmt.Errorf("dirty at a migration this release does not have: %w", err)}
		}
		prev = int(p)
	}
	slog.Warn("A DB a migration failed on is forced back to the version before it, to migrate again",
		"db", name, "failed_migration", version, "forced_to_version", prev)
	return m.Force(prev)
}

// FailedError is a DB that could not be migrated to the latest migration: the DB's name, the migration it failed at
// and why
type FailedError struct {
	DB        string
	Migration uint
	Err       error
}

func (e *FailedError) Error() string {
	return fmt.Sprintf("%s migration %d failed: %v", e.DB, e.Migration, e.Err)
}

func (e *FailedError) Unwrap() error {
	return e.Err
}

// Version is the migration db is at, 0 before its first, and whether a migration failed there
func Version(db *sql.DB, fsys fs.FS) (uint, bool, error) {
	src, err := iofs.New(fsys, "migrations")
	if err != nil {
		return 0, false, err
	}
	driver, err := sqlite.WithInstance(db, &sqlite.Config{})
	if err != nil {
		return 0, false, err
	}
	m, err := migrate.NewWithInstance("iofs", src, "sqlite", driver)
	if err != nil {
		return 0, false, err
	}
	version, dirty, err := m.Version()
	if errors.Is(err, migrate.ErrNilVersion) {
		return 0, false, nil
	}
	return version, dirty, err
}

// Latest is the last of the migrations in fsys's migrations folder
func Latest(fsys fs.FS) (uint, error) {
	src, err := iofs.New(fsys, "migrations")
	if err != nil {
		return 0, err
	}
	latest, err := src.First()
	for err == nil {
		next, nextErr := src.Next(latest)
		if nextErr != nil {
			break
		}
		latest = next
	}
	return latest, err
}
