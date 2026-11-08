package sitesettings_test

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/models/sitemodels"
	"nighttroop.t1d.nz/server/components/site_templ/site_settings_templ"
	"nighttroop.t1d.nz/server/siteapi/siteapiutils"
	"nighttroop.t1d.nz/server/siteapi/sitecollections/sitesettings"
	"nighttroop.t1d.nz/utils/consts"
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

// A site without a connector of a type has none to get, which is how saving one creates it, so it is not an error
func TestGetConnectorTheSiteHasNotGot(t *testing.T) {
	cfg, controlDB, site := testutils.NewControlDBWithSite(t)
	ctx := httputils.CreateSiteContext(context.Background(), cfg, controlDB, site.SiteDB, nil, site.SiteDB.Subdomain, "192.0.2.1", httputils.HTTPActionWrite, httputils.AcceptTypeJSON, httputils.ContentTypeJSON)
	tx := siteapiutils.NewCollectionTx(ctx, httputils.GetSiteContext(ctx))
	defer tx.Rollback()

	logged := &levels{}
	defaultLogger := slog.Default()
	slog.SetDefault(slog.New(logged))
	t.Cleanup(func() { slog.SetDefault(defaultLogger) })

	source, err := (&sitesettings.ConnectorSettingsCollection{}).GetRow(ctx, tx, "nightscout", httputils.GetSiteContext(ctx))
	require.NoError(t, err)
	require.Nil(t, source)
	for _, level := range logged.got {
		require.Less(t, level, slog.LevelError, "nothing is logged as an error, which goes to Sentry")
	}
}

// Virtual connector, which makes up glucose, is offered and saved only on a development, test or staging server, never
// in production, where it could be taken for real data
func TestVirtualConnectorIsOnlyOnDevelopmentTestAndStaging(t *testing.T) {
	tests := []struct {
		mode    consts.ServerMode
		allowed bool
	}{
		{consts.Development, true},
		{consts.Test, true},
		{consts.Staging, true},
		{consts.Production, false},
	}
	for _, test := range tests {
		t.Run(string(test.mode), func(t *testing.T) {
			testCfg, controlDB, site := testutils.NewControlDBWithSite(t)
			cfg := *testCfg
			cfg.Mode = test.mode
			ctx := httputils.CreateSiteContext(context.Background(), &cfg, controlDB, site.SiteDB, nil, site.SiteDB.Subdomain, "192.0.2.1", httputils.HTTPActionWrite, httputils.AcceptTypeJSON, httputils.ContentTypeJSON)
			siteContext := httputils.GetSiteContext(ctx)
			tx := siteapiutils.NewCollectionTx(ctx, siteContext)
			defer tx.Rollback()
			collection := &sitesettings.ConnectorSettingsCollection{}
			virtual := &sitemodels.Connector{Type: consts.ConnectorTypeVirtual, Enabled: true}

			_, createErr := collection.CreateRow(ctx, tx, new(string(consts.ConnectorTypeVirtual)), siteContext, virtual)
			_, updateErr := collection.UpdateRow(ctx, tx, siteContext, virtual, virtual)
			var page bytes.Buffer
			require.NoError(t, site_settings_templ.ConnectorSettingsPartial(nil).Render(ctx, &page))
			offered := strings.Contains(page.String(), "site--settings--virtual-connector--enable")
			if test.allowed {
				require.NoError(t, createErr)
				require.NoError(t, updateErr)
				require.True(t, offered)
				return
			}
			require.ErrorContains(t, createErr, "virtual connector, which makes up glucose, is only on development, test and staging servers")
			require.ErrorContains(t, updateErr, "virtual connector, which makes up glucose, is only on development, test and staging servers")
			require.False(t, offered)
			// The other sources are still offered
			require.Contains(t, page.String(), "site--settings--nightscout-connector--enable")
		})
	}
}
