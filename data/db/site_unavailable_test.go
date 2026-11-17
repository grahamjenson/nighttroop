package db_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/utils/consts"
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
	messages []string
	args     [][]any
}

func (p *pages) page(_ context.Context, message string, args ...any) {
	p.mu.Lock()
	defer p.mu.Unlock()
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
