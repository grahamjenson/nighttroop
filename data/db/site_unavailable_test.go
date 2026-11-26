package db_test

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"io"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/page"
	"nighttroop.t1d.nz/utils/testutils"
)

// missingSite is a control DB on disk, opened again after its site's DB file is gone, so the site is unavailable, and a
// copy of the DB as its backup
func missingSite(t *testing.T) (*db.ControlDB, string, []byte) {
	cfg := testutils.GetTestServerConfig()
	cfg.Mode, cfg.DataDir, cfg.Status.S3Working = consts.Development, t.TempDir(), true
	controlDB, err := db.NewControlDB(t.Context(), cfg)
	require.NoError(t, err)
	_, err = controlDB.CreateSiteWithID(t.Context(), "kostsite")
	require.NoError(t, err)
	require.NoError(t, controlDB.Close())
	path := sitedb.GetPath("kostsite", cfg)
	backup, err := os.ReadFile(path)
	require.NoError(t, err)
	require.NoError(t, os.Remove(path))
	controlDB, err = db.NewControlDB(t.Context(), cfg)
	require.NoError(t, err)
	t.Cleanup(func() { controlDB.Close() })
	require.Equal(t, map[string]string{"kostsite": db.UnavailableMissing}, controlDB.UnavailableSites())
	return controlDB, "kostsite", backup
}

// pages records each page
type pages struct {
	mu       sync.Mutex
	levels   []page.Level
	messages []string
	args     [][]any
}

func (p *pages) page(_ context.Context, level page.Level, message string, args ...any) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.levels = append(p.levels, level)
	p.messages = append(p.messages, message)
	p.args = append(p.args, args)
}

// An unavailable site whose restore fails is tried again after each wait, and once restored is served, with no page;
// one never restored is tried after every wait, then the admin is paged once, naming it
func TestKeepRestoringUnavailableSites(t *testing.T) {
	retries := []time.Duration{time.Millisecond, time.Millisecond, time.Millisecond}
	tests := []struct {
		name string
		// failures is how many tries fail before the download works
		failures int
		restored bool
	}{
		{name: "restored on its third try", failures: 2, restored: true},
		{name: "never restored, the admin is paged", failures: 10},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			controlDB, siteID, backup := missingSite(t)
			tries := 0
			download := func(_ context.Context, _ string, outpath string) (int64, bool, error) {
				tries++
				if tries <= tt.failures {
					return 0, false, errors.New("S3 is down")
				}
				require.NoError(t, os.MkdirAll(outpath, 0o755))
				return int64(len(backup)), true, os.WriteFile(filepath.Join(outpath, siteID+".sqlite"), backup, 0o600)
			}
			paged := &pages{}
			controlDB.KeepRestoringUnavailableSites(t.Context(), download, retries, paged.page)

			_, _, err := controlDB.GetSite(t.Context(), siteID)
			if tt.restored {
				require.NoError(t, err)
				require.Equal(t, tt.failures+1, tries, "tried until it is restored")
				require.Empty(t, paged.messages)
				return
			}
			require.ErrorIs(t, err, db.ErrSiteUnavailable)
			require.Equal(t, len(retries)+1, tries, "tried once, then after each wait")
			require.Equal(t, []page.Level{page.Call}, paged.levels, "sites down call the admin")
			require.Equal(t, []string{"Sites are unavailable, as their DBs could not be restored"}, paged.messages)
			require.Equal(t, []any{"site_ids", []string{siteID}, "tries", len(retries) + 1}, paged.args[0])
		})
	}
}

// A server that stops while waiting to try again stops trying, and pages no one
func TestKeepRestoringStopsWithTheServer(t *testing.T) {
	controlDB, _, _ := missingSite(t)
	ctx, stop := context.WithCancel(t.Context())
	tries := 0
	download := func(context.Context, string, string) (int64, bool, error) {
		tries++
		stop()
		return 0, false, errors.New("S3 is down")
	}
	paged := &pages{}
	controlDB.KeepRestoringUnavailableSites(ctx, download, []time.Duration{time.Hour}, paged.page)
	require.Equal(t, 1, tries)
	require.Empty(t, paged.messages)
}

// A deleted site that is unavailable, as its DB did not open, is not restored any more and loses the DB files it has.
// The sites deleted with it are still closed and lose theirs.
func TestDeleteSitesWithAnUnavailableSite(t *testing.T) {
	controlDB, siteID, _ := missingSite(t)
	ctx := t.Context()
	path := sitedb.GetPath(siteID, controlDB.ServerConfig)
	require.NoError(t, os.WriteFile(path, []byte("not a DB"), 0o600))
	other, err := controlDB.CreateSiteWithID(ctx, "opensite")
	require.NoError(t, err)

	require.NoError(t, controlDB.DeleteSites(ctx, []string{siteID, "opensite"}))
	require.Empty(t, controlDB.UnavailableSites())
	require.NoFileExists(t, path)
	require.NoFileExists(t, other.SiteDB.Path)
	tries := 0
	controlDB.RestoreUnavailableSites(ctx, func(context.Context, string, string) (int64, bool, error) {
		tries++
		return 0, false, nil
	})
	require.Zero(t, tries, "a deleted site is not restored")
}

// A control admin makes an unavailable site's DB again, empty, so it is served again, and counts as written, so its new
// DB is backed up. A site that is not unavailable, or whose DB is being restored now, is not made again.
func TestNewEmptySiteDB(t *testing.T) {
	controlDB, siteID, _ := missingSite(t)
	ctx := t.Context()

	// While its restore is downloading, the site is not made again
	controlDB.RestoreUnavailableSites(ctx, func(context.Context, string, string) (int64, bool, error) {
		require.ErrorIs(t, controlDB.NewEmptySiteDB(ctx, siteID), db.ErrSiteNotUnavailable)
		return 0, false, nil
	})

	require.NoError(t, controlDB.NewEmptySiteDB(ctx, siteID))
	site, model, err := controlDB.GetSite(ctx, siteID)
	require.NoError(t, err, "it is served again")
	require.Empty(t, controlDB.UnavailableSites())
	require.NotNil(t, model.LastWriteAt, "it counts as written, so it is backed up")
	readings, err := site.Queries().SearchGlucoseReadings(ctx, sitedb.SearchGlucoseReadingsParams{Limit: 10})
	require.NoError(t, err)
	require.Empty(t, readings, "its DB is empty")

	require.ErrorIs(t, controlDB.NewEmptySiteDB(ctx, siteID), db.ErrSiteNotUnavailable, "an available site is not made again")
}

// A site whose DB did not open keeps that DB beside its new, empty one
func TestNewEmptySiteDBKeepsTheDBThatDidNotOpen(t *testing.T) {
	controlDB, siteID, _ := missingSite(t)
	path := sitedb.GetPath(siteID, controlDB.ServerConfig)
	require.NoError(t, os.WriteFile(path, []byte("not a DB"), 0o600))
	controlDB.MakeSiteUnavailable(siteID, db.UnavailableCannotOpen)

	require.NoError(t, controlDB.NewEmptySiteDB(t.Context(), siteID))
	kept, err := filepath.Glob(path + ".unopened-*")
	require.NoError(t, err)
	require.Len(t, kept, 1)
	data, err := os.ReadFile(kept[0])
	require.NoError(t, err)
	require.Equal(t, "not a DB", string(data))
}

// A control admin uploads a fixed DB in place of an unavailable site's, which it is served from again, counted as
// written so it is backed up. One that does not open leaves the site unavailable, with its DB there put back, and a
// site that is not unavailable has none put in place.
func TestUploadSiteDB(t *testing.T) {
	controlDB, siteID, backup := missingSite(t)
	ctx := t.Context()
	path := sitedb.GetPath(siteID, controlDB.ServerConfig)
	require.NoError(t, os.WriteFile(path, []byte("not a DB"), 0o600))
	controlDB.MakeSiteUnavailable(siteID, db.UnavailableCannotOpen)

	err := controlDB.UploadSiteDB(ctx, siteID, strings.NewReader("not a DB either"))
	require.ErrorIs(t, err, db.ErrDBDoesNotOpen)
	require.Contains(t, controlDB.UnavailableSites(), siteID)
	kept, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, "not a DB", string(kept), "its DB is put back")

	require.NoError(t, controlDB.UploadSiteDB(ctx, siteID, bytes.NewReader(backup)))
	_, model, err := controlDB.GetSite(ctx, siteID)
	require.NoError(t, err, "it is served again")
	require.Empty(t, controlDB.UnavailableSites())
	require.NotNil(t, model.LastWriteAt, "it counts as written, so it is backed up")
	aside, err := filepath.Glob(path + ".unopened-*")
	require.NoError(t, err)
	require.Len(t, aside, 1, "the DB that did not open is kept beside it")

	require.ErrorIs(t, controlDB.UploadSiteDB(ctx, siteID, bytes.NewReader(backup)), db.ErrSiteNotUnavailable)
}

// A control admin restores an unavailable site from a backup they picked, which it is served from again. A backup not
// there leaves it unavailable, and a site that is not unavailable is not restored.
func TestRestorePickedBackup(t *testing.T) {
	controlDB, siteID, backup := missingSite(t)
	ctx := t.Context()
	notThere := func(context.Context, string, string) (int64, bool, error) { return 0, false, nil }
	restored, err := controlDB.RestoreUnavailableSite(ctx, siteID, notThere)
	require.NoError(t, err)
	require.False(t, restored)
	_, _, err = controlDB.GetSite(ctx, siteID)
	require.ErrorIs(t, err, db.ErrSiteUnavailable)

	picked := func(_ context.Context, _ string, outpath string) (int64, bool, error) {
		require.NoError(t, os.MkdirAll(outpath, 0o755))
		return int64(len(backup)), true, os.WriteFile(filepath.Join(outpath, siteID+".sqlite"), backup, 0o600)
	}
	restored, err = controlDB.RestoreUnavailableSite(ctx, siteID, picked)
	require.NoError(t, err)
	require.True(t, restored)
	_, _, err = controlDB.GetSite(ctx, siteID)
	require.NoError(t, err, "it is served again")
	require.Empty(t, controlDB.UnavailableSites())

	_, err = controlDB.RestoreUnavailableSite(ctx, siteID, picked)
	require.ErrorIs(t, err, db.ErrSiteNotUnavailable)
}

// A control admin downloads the DB here of an unavailable site, such as one whose migration failed, as it is, to fix it.
// A site with no DB here, or whose DB is being restored, has none to download.
func TestSnapshotUnavailableSite(t *testing.T) {
	controlDB, siteID, backup := missingSite(t)
	ctx := t.Context()
	_, err := controlDB.SnapshotUnavailableSite(ctx, siteID)
	require.ErrorIs(t, err, db.ErrNoSiteDBHere)

	path := sitedb.GetPath(siteID, controlDB.ServerConfig)
	require.NoError(t, os.WriteFile(path, backup, 0o600))
	controlDB.MakeSiteUnavailable(siteID, db.UnavailableMigrationFailed)
	controlDB.RestoreUnavailableSites(ctx, func(context.Context, string, string) (int64, bool, error) {
		_, err := controlDB.SnapshotUnavailableSite(ctx, siteID)
		require.ErrorIs(t, err, db.ErrSiteNotUnavailable, "not while it is being restored")
		return 0, false, nil
	})

	download, err := controlDB.SnapshotUnavailableSite(ctx, siteID)
	require.NoError(t, err)
	defer download.Close()
	w := httptest.NewRecorder()
	require.NoError(t, download.Send(w))
	archive, err := zip.NewReader(bytes.NewReader(w.Body.Bytes()), int64(w.Body.Len()))
	require.NoError(t, err)
	require.Len(t, archive.File, 1)
	entry, err := archive.File[0].Open()
	require.NoError(t, err)
	copied, err := io.ReadAll(entry)
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(string(copied), "SQLite format 3\x00"), "the zip holds a SQLite DB")

	kept, err := os.ReadFile(path)
	require.NoError(t, err)
	require.Equal(t, backup, kept, "its DB is not changed")
	require.Contains(t, controlDB.UnavailableSites(), siteID, "it stays unavailable")
}
