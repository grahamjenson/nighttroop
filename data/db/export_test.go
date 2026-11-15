package db

import (
	"context"

	"nighttroop.t1d.nz/utils/aslog"
)

// FlushFrom writes buffer's held logs, for a test to write its own buffer's rather than the server's shared one
func (l *LogsDB) FlushFrom(buffer *aslog.Buffer) { l.flush(buffer) }

// OpenSite puts site's DB with the server's open ones, for a test to stand in a DB on disk for one in memory
func (db *ControlDB) OpenSite(site *SiteDB) { db.sites.add(site) }

// TruncateSiteSince truncates site as the job does, looking at its device reports written since since
func TruncateSiteSince(ctx context.Context, site *SiteDB, since int64) error { return truncateSite(ctx, site, since) }

// StartSetupIfNoAdmin checks again, as the server does as it starts, whether it has an admin or is to be set up
func (db *ControlDB) StartSetupIfNoAdmin(ctx context.Context) error { return db.startSetupIfNoAdmin(ctx) }

// SetSizeKB makes sizeKB the site DB's size, for a test of a DB in memory
func (db *SiteDB) SetSizeKB(sizeKB func() *int64) { db.sizeKB = sizeKB }
