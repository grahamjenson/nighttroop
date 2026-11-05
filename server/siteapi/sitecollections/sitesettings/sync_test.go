package sitesettings_test

import (
	"context"
	"log/slog"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/server/siteapi/siteapiutils"
	"nighttroop.t1d.nz/server/siteapi/sitecollections/sitesettings"
	"nighttroop.t1d.nz/utils/httputils"
	"nighttroop.t1d.nz/utils/testutils"
)

// levels records the level of each log
type levels struct {
	mu  sync.Mutex
	got []slog.Level
}

func (l *levels) Enabled(context.Context, slog.Level) bool { return true }
func (l *levels) Handle(_ context.Context, record slog.Record) error {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.got = append(l.got, record.Level)
	return nil
}
func (l *levels) WithAttrs([]slog.Attr) slog.Handler { return l }
func (l *levels) WithGroup(string) slog.Handler      { return l }

// A site without a sync source of a type has none to get, which is how saving one creates it, so it is not an error
func TestGetSyncSourceTheSiteHasNotGot(t *testing.T) {
	cfg, controlDB, site := testutils.NewControlDBWithSite(t)
	ctx := httputils.CreateSiteContext(context.Background(), cfg, controlDB, site.SiteDB, nil, site.SiteDB.Subdomain, "192.0.2.1", httputils.HTTPActionWrite, httputils.AcceptTypeJSON, httputils.ContentTypeJSON)
	tx := siteapiutils.NewCollectionTx(ctx, httputils.GetSiteContext(ctx))
	defer tx.Rollback()

	logged := &levels{}
	defaultLogger := slog.Default()
	slog.SetDefault(slog.New(logged))
	t.Cleanup(func() { slog.SetDefault(defaultLogger) })

	source, err := (&sitesettings.SyncSettingsCollection{}).GetRow(ctx, tx, "nightscout", httputils.GetSiteContext(ctx))
	require.NoError(t, err)
	require.Nil(t, source)
	for _, level := range logged.got {
		require.Less(t, level, slog.LevelError, "nothing is logged as an error, which goes to Sentry")
	}
}
