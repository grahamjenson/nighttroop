package db

import (
	"context"
	"fmt"

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

// WARNING THIS IS A BEST TRY FUNCTION ALL ERRORS ARE SWALLOWED

func (db *SiteDB) URL() string {
	return fmt.Sprintf("%s://%s.%s:%d", db.ServerConfig.HostProtocol, db.Subdomain, db.ServerConfig.HostName, db.ServerConfig.HostPort)
}

func (db *SiteDB) NSURL() string {
	return fmt.Sprintf("%s://%s.ns.%s:%d", db.ServerConfig.HostProtocol, db.Subdomain, db.ServerConfig.HostName, db.ServerConfig.HostPort)
}
