package db

import (
	"context"
	"database/sql"
	"fmt"

	"nighttroop.t1d.nz/data/sql/site"

	"nighttroop.t1d.nz/utils"
	"nighttroop.t1d.nz/utils/config"
)

// //
// TenentDB
// //

type SiteDB struct {
	serverConfig config.ServerConfig

	SQLiteDB  *sql.DB
	Subdomain string

	// Prepared statements
	Queries *site.Queries
}

func NewSiteDB(ctx context.Context, subdomain string, mustExist bool, cfg config.ServerConfig) (*SiteDB, error) {
	path, _ := utils.SiteDBPaths(subdomain)
	if mustExist && !utils.FileExists(path) {
		return nil, fmt.Errorf("The Site DB %s Must exist and doesnt", subdomain)
	}

	sqldb, err := OpenSqliteDB(path, 100, cfg.IsTest())
	if err != nil {
		return nil, fmt.Errorf("cannot open site db: %s %w", subdomain, err)
	}

	// Create a new site db
	siteDB := &SiteDB{
		serverConfig: cfg,
		SQLiteDB:     sqldb,
		Subdomain:    subdomain,
	}

	err = site.MigrateUp(subdomain, sqldb)
	if err != nil {
		sqldb.Close()
		return siteDB, fmt.Errorf("Migration Error Occurred site db: %s %w", subdomain, err)
	}

	siteDB.Queries, err = site.Prepare(context.TODO(), sqldb)
	if err != nil {
		sqldb.Close()
		return siteDB, fmt.Errorf("cannot prepare statements for site db: %s %w", subdomain, err)
	}

	return siteDB, nil
}

// WARNING THIS IS A BEST TRY FUNCTION ALL ERRORS ARE SWOLLOWED
func (db *SiteDB) TryCloseAndDelete() {
	if db == nil {
		return
	}
	if db.SQLiteDB != nil {
		db.SQLiteDB.Close()
	}
	if db.Subdomain == "" {
		return
	}
	path, _ := utils.SiteDBPaths(db.Subdomain)
	_ = utils.DeleteFile(path)
}

func (db *SiteDB) Close() error {
	return db.SQLiteDB.Close()
}

func (db *SiteDB) URL() string {
	return fmt.Sprintf("%s://%s.%s:%d", db.serverConfig.HostProtocol, db.Subdomain, db.serverConfig.HostName, db.serverConfig.HostPort)
}

func (db *SiteDB) NSURL() string {
	return fmt.Sprintf("%s://%s.ns.%s:%d", db.serverConfig.HostProtocol, db.Subdomain, db.serverConfig.HostName, db.serverConfig.HostPort)
}
