package db

import (
	"context"
	"fmt"

	"nighttroop.t1d.nz/data/sql/site"

	"nighttroop.t1d.nz/utils"
	"nighttroop.t1d.nz/utils/config"
)

// //
// TenentDB
// //

type SiteDB struct {
	*site.DB

	Subdomain string
}

func NewSiteDB(ctx context.Context, subdomain string, cfg config.ServerConfig) (*SiteDB, error) {
	path, _ := utils.SiteDBPaths(subdomain)
	rawDB, err := site.NewDB(ctx, path, 100, cfg)
	if err != nil {
		return nil, fmt.Errorf("cannot open site db: %s %w", subdomain, err)
	}

	// Create a new site db
	siteDB := &SiteDB{
		DB:        rawDB,
		Subdomain: subdomain,
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
	return fmt.Sprintf("%s://%s.%s:%d", db.ServerConfig.HostProtocol, db.Subdomain, db.ServerConfig.HostName, db.ServerConfig.HostPort)
}

func (db *SiteDB) NSURL() string {
	return fmt.Sprintf("%s://%s.ns.%s:%d", db.ServerConfig.HostProtocol, db.Subdomain, db.ServerConfig.HostName, db.ServerConfig.HostPort)
}
