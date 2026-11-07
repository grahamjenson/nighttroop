// sites fixes a start that a site stopped, run in the server's image against its data while the server is stopped:
//
//	docker compose run --rm --no-deps --entrypoint sites nighttroop check
//
// check lists what stops the start: the control DB, or a live or disabled site's DB, that cannot be opened or
// migrated, and a site whose DB is missing, which the start restores only from its batch's backup in S3, each with
// its last backup in the data folder. It changes nothing, migrating a copy of each DB, and exits 1 when something
// stops the start.
package main

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/utils"
	"nighttroop.t1d.nz/utils/backup"
	"nighttroop.t1d.nz/utils/config"
)

func main() {
	stopped, err := run(context.Background())
	if err != nil {
		fmt.Fprintln(os.Stderr, "sites:", err)
	}
	if err != nil || stopped {
		os.Exit(1)
	}
}

// run runs the command named after the server's flags, which its config reads
func run(ctx context.Context) (bool, error) {
	cfg, err := config.NewServerConfig()
	if err != nil {
		return false, err
	}
	if flag.Arg(0) != "check" {
		return false, errors.New("usage: sites check")
	}
	return check(ctx, cfg, os.Stdout)
}

// check writes to out what stops a server with cfg's data starting, with each one's last backup, and is whether
// anything does
func check(ctx context.Context, cfg *config.ServerConfig, out io.Writer) (bool, error) {
	tmp, err := os.MkdirTemp("", "sites-check-")
	if err != nil {
		return false, err
	}
	defer os.RemoveAll(tmp)

	if !utils.FileExists(controldb.GetPath(cfg)) {
		fmt.Fprintln(out, "There is no control DB: the start restores it from S3, or sets up a new server")
		return false, nil
	}
	control, err := openCopy(ctx, controldb.GetPath(cfg), filepath.Join(tmp, "controldb.sqlite"), controldb.OpenSqliteDB, controldb.MigrateUp)
	if err != nil {
		report(out, "control DB", err.Error(), controldb.GetBackupPath(cfg))
		return true, nil
	}
	defer control.Close()
	sites, err := controldb.New(control).GetLiveAndDisabledSites(ctx)
	if err != nil {
		return false, fmt.Errorf("cannot list the sites: %w", err)
	}

	stopped := false
	for _, site := range sites {
		why := "its DB is missing, so the start stops unless its batch's backup in S3 has it"
		if path := sitedb.GetPath(site.ID, cfg); utils.FileExists(path) {
			err := checkSite(ctx, path, tmp)
			if err == nil {
				continue
			}
			why = err.Error()
		}
		stopped = true
		report(out, fmt.Sprintf("site %s (%s)", site.ID, site.State), why, sitedb.GetBackupPath(site.ID, cfg))
	}
	if !stopped {
		fmt.Fprintf(out, "Every one of the %d live and disabled sites opens and migrates\n", len(sites))
	}
	return stopped, nil
}

// checkSite opens and migrates a copy in tmp of the site DB at path, then removes it, so there is one copy at a time
func checkSite(ctx context.Context, path, tmp string) error {
	copyPath := filepath.Join(tmp, filepath.Base(path))
	db, err := openCopy(ctx, path, copyPath, sitedb.OpenSqliteDB, sitedb.MigrateUp)
	if err == nil {
		db.Close()
	}
	return errors.Join(err, utils.DeleteFiles([]string{copyPath, copyPath + "-wal", copyPath + "-shm"}))
}

// openCopy copies the DB at path, opened read only, to copyPath, and opens and migrates the copy as the server opens
// the DB, so the DB is not changed
func openCopy(ctx context.Context, path, copyPath string, open func(string, int) (*sql.DB, error), migrate func(*sql.DB) error) (*sql.DB, error) {
	src, err := sql.Open("sqlite", "file:"+path+"?mode=ro")
	if err == nil {
		defer src.Close()
		err = backup.SqliteBackup(ctx, src, copyPath)
	}
	if err != nil {
		return nil, fmt.Errorf("cannot open: %w", err)
	}
	db, err := open(copyPath, 0)
	if err != nil {
		return nil, err
	}
	if err := migrate(db); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

// report writes what stops the start, why, and when its backup in the data folder was made
func report(out io.Writer, name, why, backupPath string) {
	local := "none"
	if info, err := os.Stat(backupPath); err == nil {
		local = backupPath + ", made " + info.ModTime().UTC().Format(time.RFC3339)
	}
	fmt.Fprintf(out, "%s: %s\n  backup in the data folder: %s\n", name, why, local)
}
