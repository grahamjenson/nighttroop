package db

import (
	"context"
	"fmt"
	"os"

	"nighttroop.t1d.nz/data/sql/sitedb"

	"nighttroop.t1d.nz/utils/config"
)

// //
// TenentDB
// //

type SiteDB struct {
	*sitedb.DB

	Subdomain string
}

func NewSiteDB(ctx context.Context, subdomain string, cfg *config.ServerConfig) (*SiteDB, error) {
	rawDB, err := sitedb.NewDB(ctx, subdomain, cfg)
	if err != nil {
		// NewDB returns the DB when it opened but failed later, so the half made file can be removed
		if rawDB != nil {
			rawDB.TryCloseAndDelete()
		}
		return nil, fmt.Errorf("cannot open site db: %s %w", subdomain, err)
	}

	// Create a new site db
	siteDB := &SiteDB{
		DB:        rawDB,
		Subdomain: subdomain,
	}

	return siteDB, nil
}

// EstimatedSizeKB is the site DB's size on disk in KB, its file and write-ahead log, or nil for one in memory
func (db *SiteDB) EstimatedSizeKB() *int64 {
	if db.InMemory {
		return nil
	}
	var size int64
	for _, path := range []string{db.Path, db.Path + "-wal"} {
		if info, err := os.Stat(path); err == nil {
			size += info.Size()
		}
	}
	kb := size / 1024
	return &kb
}

// WARNING THIS IS A BEST TRY FUNCTION ALL ERRORS ARE SWALLOWED

func (db *SiteDB) URL() string {
	return fmt.Sprintf("%s://%s.%s:%d", db.ServerConfig.HostProtocol, db.Subdomain, db.ServerConfig.HostName, db.ServerConfig.HostPort)
}

func (db *SiteDB) NSURL() string {
	return fmt.Sprintf("%s://%s.ns.%s:%d", db.ServerConfig.HostProtocol, db.Subdomain, db.ServerConfig.HostName, db.ServerConfig.HostPort)
}
