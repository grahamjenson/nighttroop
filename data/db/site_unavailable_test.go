package db_test

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/data/sql/logsdb"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/s3utils"
	"nighttroop.t1d.nz/utils/testutils"
)

// fileServer is a config whose DBs are files in a folder of their own, as on a server, with S3 working or not
func fileServer(t *testing.T, s3Working bool) *config.ServerConfig {
	cfg := *testutils.GetTestServerConfig()
	cfg.Mode, cfg.DataDir, cfg.Status.S3Working = consts.Development, t.TempDir(), s3Working
	return &cfg
}

// restart opens the control DB again, as a start does, with its tries to restore a site an hour apart unless backoff
// is given, counting its pages
func restart(t *testing.T, cfg *config.ServerConfig, pages *atomic.Int32, backoff ...time.Duration) *db.ControlDB {
	controlDB, err := db.NewControlDB(t.Context(), cfg)
	require.NoError(t, err, "the server starts")
	t.Cleanup(func() { controlDB.Close() })
	controlDB.RestoreBackoff = []time.Duration{time.Hour}
	if len(backoff) > 0 {
		controlDB.RestoreBackoff = backoff
	}
	controlDB.Page = func(context.Context, string, ...any) { pages.Add(1) }
	return controlDB
}

// newSites makes the sites, then closes the control DB, as a server stopping
func newSites(t *testing.T, cfg *config.ServerConfig, siteIDs ...string) {
	controlDB, err := db.NewControlDB(t.Context(), cfg)
	require.NoError(t, err)
	for _, siteID := range siteIDs {
		_, err := controlDB.CreateSiteWithID(t.Context(), siteID)
		require.NoError(t, err)
	}
	require.NoError(t, controlDB.Close())
}

// backupOf moves the site's DB out, as its backup in a stand in for S3, which download then gives, or not found once
// missing is set. Like s3utils.DownloadFilesFromS3 it makes its staging folder beside outpath, and leaves it there, as
// a start killed part way through the download does.
func backupOf(t *testing.T, cfg *config.ServerConfig, siteID string, missing *atomic.Bool) s3utils.DownloadFunc {
	backup := filepath.Join(t.TempDir(), siteID+".sqlite")
	require.NoError(t, os.Rename(sitedb.GetPath(siteID, cfg), backup))
	return func(_ context.Context, key string, outpath string) (int64, bool, error) {
		if missing.Load() || key != sitedb.GetS3SitePath(siteID, cfg) {
			return 0, false, nil
		}
		_, err := os.MkdirTemp(filepath.Dir(outpath), "."+filepath.Base(outpath)+"-restore-")
		require.NoError(t, err)
		require.NoError(t, os.MkdirAll(outpath, 0o755))
		data, err := os.ReadFile(backup)
		require.NoError(t, err)
		return int64(len(data)), true, os.WriteFile(filepath.Join(outpath, siteID+".sqlite"), data, 0o600)
	}
}

// stateOf is the site's state and why it is unavailable
func stateOf(t *testing.T, controlDB *db.ControlDB, siteID string) (string, string) {
	row, err := controlDB.Queries.GetSiteInAnyState(t.Context(), controldb.GetSiteInAnyStateParams{ID: siteID})
	require.NoError(t, err)
	if row.UnavailableReason == nil {
		return row.State, ""
	}
	return row.State, *row.UnavailableReason
}

// restoreEvents are the attributes of the site's site_unavailable and site_available events, oldest first
func restoreEvents(t *testing.T, controlDB *db.ControlDB, siteID string) []string {
	found, err := controlDB.Logs.Queries.SearchEvents(t.Context(), logsdb.SearchEventsParams{SiteID: &siteID, Limit: 100})
	require.NoError(t, err)
	events := []string{}
	for _, event := range found {
		if event.Name == "site_unavailable" || event.Name == "site_available" {
			events = append([]string{event.Name + " " + *event.Attributes}, events...)
		}
	}
	return events
}

// A live site whose DB file is missing at start, and one whose DB does not open, are unavailable, and the server
// starts with the rest: the DB that does not open is kept, moved aside, its requests are told the site is
// unavailable, and a disabled site's that it is not there. The control admin sees each one.
func TestStartWithSitesUnavailable(t *testing.T) {
	cfg, pages := fileServer(t, false), &atomic.Int32{}
	newSites(t, cfg, "keptsite", "kostsite", "kranksit", "kidsitea")
	require.NoError(t, os.Remove(sitedb.GetPath("kostsite", cfg)))
	require.NoError(t, os.WriteFile(sitedb.GetPath("kranksit", cfg), []byte("not a database"), 0o600))
	controlDB := restart(t, cfg, pages)
	_, err := controlDB.Queries.DisableSite(t.Context(), controldb.DisableSiteParams{ID: "kidsitea"})
	require.NoError(t, err)
	require.NoError(t, controlDB.Close())
	require.NoError(t, os.Remove(sitedb.GetPath("kidsitea", cfg)))

	controlDB = restart(t, cfg, pages)
	_, _, err = controlDB.GetSite(t.Context(), "keptsite")
	require.NoError(t, err)
	for siteID, reason := range map[string]string{"kostsite": db.UnavailableMissing, "kranksit": db.UnavailableCannotOpen} {
		_, _, err = controlDB.GetSite(t.Context(), siteID)
		require.ErrorIs(t, err, db.ErrSiteUnavailable, siteID)
		state, why := stateOf(t, controlDB, siteID)
		require.Equal(t, []string{"unavailable", reason}, []string{state, why}, siteID)
	}
	_, _, err = controlDB.GetSite(t.Context(), "kidsitea")
	require.ErrorIs(t, err, db.ErrSiteNotFound, "a disabled site is not there, unavailable or not")
	broken, err := filepath.Glob(sitedb.GetPath("kranksit", cfg) + ".broken-*")
	require.NoError(t, err)
	require.Len(t, broken, 1)
	data, err := os.ReadFile(broken[0])
	require.NoError(t, err)
	require.Equal(t, "not a database", string(data), "the DB that does not open is kept")
	require.NoFileExists(t, sitedb.GetPath("kranksit", cfg))

	ids := []string{}
	for _, site := range controlDB.UnavailableSites() {
		ids = append(ids, site.ID)
	}
	require.ElementsMatch(t, []string{"kostsite", "kranksit", "kidsitea"}, ids)
	sites, err := controlDB.Queries.GetSitesToOpen(t.Context())
	require.NoError(t, err)
	require.Len(t, sites, 4)
	require.Len(t, controlDB.GetSiteDBsFromModels(sites), 1, "only the site whose DB opens is open")
}

// A start tries once to restore each unavailable site from its own backup, with its data, then goes on: a site
// restored is live again, or disabled if it was, and one that is not stays unavailable for why. Each try is an event
// the control admin sees.
func TestRestoreUnavailableSites(t *testing.T) {
	cfg, pages, missing := fileServer(t, true), &atomic.Int32{}, &atomic.Bool{}
	newSites(t, cfg, "keptsite", "kostsite", "kingsite")
	controlDB := restart(t, cfg, pages)
	site, _, err := controlDB.GetSite(t.Context(), "kostsite")
	require.NoError(t, err)
	_, err = site.Queries().CreateGlucoseReading(t.Context(), sitedb.CreateGlucoseReadingParams{ID: "kept", Mgdl: 120, ReadingType: "sgv", OccurredAt: 1, CreatedByIdentityName: "phone"})
	require.NoError(t, err)
	require.NoError(t, controlDB.Close())
	download := backupOf(t, cfg, "kostsite", missing)
	require.NoError(t, os.Remove(sitedb.GetPath("kingsite", cfg)))

	controlDB = restart(t, cfg, pages)
	controlDB.RestoreUnavailableSites(t.Context(), download)
	site, _, err = controlDB.GetSite(t.Context(), "kostsite")
	require.NoError(t, err, "the site is restored and open")
	reading, err := site.Queries().GetGlucoseReading(t.Context(), sitedb.GetGlucoseReadingParams{ID: "kept"})
	require.NoError(t, err)
	require.Equal(t, float64(120), reading.Mgdl)
	state, _ := stateOf(t, controlDB, "kostsite")
	require.Equal(t, "live", state)
	_, _, err = controlDB.GetSite(t.Context(), "kingsite")
	require.ErrorIs(t, err, db.ErrSiteUnavailable)
	state, why := stateOf(t, controlDB, "kingsite")
	require.Equal(t, []string{"unavailable", db.UnavailableNotInBackup}, []string{state, why})
	entries, err := os.ReadDir(filepath.Dir(sitedb.GetSitesFolder("k", cfg)))
	require.NoError(t, err)
	for _, entry := range entries {
		require.False(t, strings.HasPrefix(entry.Name(), "."), "no staging folder is left behind")
	}

	require.Equal(t, []string{`site_available {"try":0}`}, restoreEvents(t, controlDB, "kostsite"))
	require.Equal(t, []string{`site_unavailable {"reason":"not_in_backup","try":0}`}, restoreEvents(t, controlDB, "kingsite"))
	require.Empty(t, restoreEvents(t, controlDB, "keptsite"))
	require.Zero(t, pages.Load())
}

// A restore without S3, or of a backup that does not open, as one whose migration fails, leaves the site unavailable,
// and the restored DB is moved aside; one restored brings a disabled site back disabled
func TestRestoreUnavailableSite(t *testing.T) {
	tests := []struct {
		name      string
		s3Working bool
		// dirty is a backup dirty at a migration it already has, so migrating it again fails, as one that keeps failing does
		dirty    bool
		disabled bool
		state    string
		reason   string
	}{
		{name: "restored", s3Working: true, state: "live"},
		{name: "restored, disabled again", s3Working: true, disabled: true, state: "disabled"},
		{name: "without S3", state: "unavailable", reason: db.UnavailableS3NotWorking},
		{name: "restored but does not open", s3Working: true, dirty: true, state: "unavailable", reason: db.UnavailableCannotOpen},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, pages, siteID := fileServer(t, tt.s3Working), &atomic.Int32{}, "kostsite"
			newSites(t, cfg, siteID)
			if tt.disabled {
				controlDB := restart(t, cfg, pages)
				_, err := controlDB.Queries.DisableSite(t.Context(), controldb.DisableSiteParams{ID: siteID})
				require.NoError(t, err)
				require.NoError(t, controlDB.Close())
			}
			if tt.dirty {
				raw, err := sql.Open("sqlite", sitedb.GetPath(siteID, cfg))
				require.NoError(t, err)
				_, err = raw.Exec("UPDATE schema_migrations SET dirty = 1")
				require.NoError(t, err)
				require.NoError(t, raw.Close())
			}
			download := backupOf(t, cfg, siteID, &atomic.Bool{})

			controlDB := restart(t, cfg, pages)
			controlDB.RestoreUnavailableSites(t.Context(), download)
			state, why := stateOf(t, controlDB, siteID)
			require.Equal(t, []string{tt.state, tt.reason}, []string{state, why})
			broken, err := filepath.Glob(sitedb.GetPath(siteID, cfg) + ".broken-*")
			require.NoError(t, err)
			require.Equal(t, tt.dirty, len(broken) == 1, "a restored DB that does not open is moved aside")
		})
	}
}

// An unavailable site is tried again after each wait, until it is restored, without paging the admin; after every
// try has failed, the admin is paged once. An admin can try it again then, which starts the tries again.
func TestRetryRestore(t *testing.T) {
	cfg, pages, missing := fileServer(t, true), &atomic.Int32{}, &atomic.Bool{}
	newSites(t, cfg, "kostsite", "keptsite")
	download := backupOf(t, cfg, "kostsite", missing)
	missing.Store(true)
	controlDB := restart(t, cfg, pages, time.Millisecond, time.Millisecond, time.Millisecond)

	controlDB.RestoreUnavailableSites(t.Context(), download)
	require.Eventually(t, func() bool { return pages.Load() == 1 }, 5*time.Second, 5*time.Millisecond, "after the last try the admin is paged")
	require.Equal(t, []db.UnavailableSite{{ID: "kostsite", Reason: db.UnavailableNotInBackup, Since: controlDB.UnavailableSites()[0].Since}}, controlDB.UnavailableSites(),
		"no tries are left")
	events := restoreEvents(t, controlDB, "kostsite")
	require.Len(t, events, 4, "the first try and 3 more")
	require.Equal(t, `site_unavailable {"reason":"not_in_backup","try":3}`, events[3])
	time.Sleep(20 * time.Millisecond)
	require.Equal(t, int32(1), pages.Load(), "the admin is paged once")

	// An admin's try, when it fails, starts the tries again, and the next restores it
	_, err := controlDB.RetryRestore(t.Context(), "keptsite", download)
	require.ErrorIs(t, err, db.ErrSiteNotFound, "only an unavailable site is restored")
	controlDB.RestoreBackoff = []time.Duration{time.Hour}
	restored, err := controlDB.RetryRestore(t.Context(), "kostsite", download)
	require.NoError(t, err)
	require.False(t, restored)
	require.Eventually(t, func() bool { return controlDB.UnavailableSites()[0].TriesLeft == 1 }, 5*time.Second, 5*time.Millisecond)
	missing.Store(false)
	restored, err = controlDB.RetryRestore(t.Context(), "kostsite", download)
	require.NoError(t, err)
	require.True(t, restored)
	require.Empty(t, controlDB.UnavailableSites())
	_, _, err = controlDB.GetSite(t.Context(), "kostsite")
	require.NoError(t, err)
	require.Equal(t, int32(1), pages.Load())
}

// A try in the background that restores the site ends the tries, and no one is paged
func TestRetryRestoreRestoresTheSite(t *testing.T) {
	cfg, pages, missing := fileServer(t, true), &atomic.Int32{}, &atomic.Bool{}
	newSites(t, cfg, "kostsite")
	download := backupOf(t, cfg, "kostsite", missing)
	missing.Store(true)
	controlDB := restart(t, cfg, pages, 50*time.Millisecond, time.Hour)

	controlDB.RestoreUnavailableSites(t.Context(), download)
	require.Len(t, controlDB.UnavailableSites(), 1)
	missing.Store(false)
	require.Eventually(t, func() bool { return len(controlDB.UnavailableSites()) == 0 }, 5*time.Second, 5*time.Millisecond)
	_, _, err := controlDB.GetSite(t.Context(), "kostsite")
	require.NoError(t, err)
	require.Equal(t, `site_available {"try":1}`, restoreEvents(t, controlDB, "kostsite")[1])
	require.Zero(t, pages.Load())
}
