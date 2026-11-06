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
	if version, dirty, verr := m.Version(); verr == nil && dirty {
		return fmt.Errorf("%s migration %d failed: %w", name, version, err)
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
			return fmt.Errorf("%s is dirty at migration %d, which is not one of its migrations: %w", name, version, err)
		}
		prev = int(p)
	}
	slog.Warn("A DB a migration failed on is forced back to the version before it, to migrate again",
		"db", name, "failed_migration", version, "forced_to_version", prev)
	return m.Force(prev)
}
