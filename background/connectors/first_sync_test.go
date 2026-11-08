package connectors_test

import (
	"bytes"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/background/connectors"
	"nighttroop.t1d.nz/clients/dexcom_client"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/data/sql/logsdb"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/errors"
	"nighttroop.t1d.nz/utils/testutils"
)

func TestFailedFirstSyncDisablesTheSource(t *testing.T) {
	cfg, controlDB, site := testutils.NewControlDBWithSite(t)
	q, ctx, siteID, typ := controlDB.Queries, t.Context(), site.SiteDB.Subdomain, consts.ConnectorTypeDexcom
	jobs := connectors.SetupConnectorJobManager(cfg, controlDB)
	enc := func(s string) *string {
		v, err := cfg.Security.Encrypt(s)
		require.NoError(t, err)
		return &v
	}
	get := func() *controldb.SiteConnector {
		s, err := q.GetConnector(ctx, controldb.GetConnectorParams{SiteID: siteID, Type: typ})
		require.NoError(t, err)
		return s
	}
	runSync := func() error { return jobs.RunConnector(ctx, connectors.ConnectorJobParams{SiteID: siteID, Type: typ}) }
	// Only sites seen in the last day are synced
	now := time.Now().UnixMilli()
	require.NoError(t, q.UpdateSiteDateFlags(ctx, controldb.UpdateSiteDateFlagsParams{LastSeenAt: &now, ID: siteID}))
	due := func() bool {
		sources, err := q.GetConnectorsToRun(ctx)
		require.NoError(t, err)
		for _, s := range sources {
			if s.SiteID == siteID && s.Type == typ {
				return true
			}
		}
		return false
	}

	// A new source with bad details: the first sync fails, turns syncing off and says why
	require.NoError(t, q.CreateConnector(ctx, controldb.CreateConnectorParams{SiteID: siteID, Type: typ, UrlOrUsernameEnc: enc("bad-user"), PasswordEnc: enc("pass"), Enabled: true}))
	require.True(t, due(), "a new source syncs straight away")
	require.Error(t, runSync())
	s := get()
	require.Nil(t, s.FirstSuccessfulSyncTime)
	require.False(t, s.Enabled)
	require.Contains(t, s.PreviousSyncErrors[0], "The first sync failed, so syncing has been turned off")
	require.False(t, due())

	// The failure is in the logs DB's errors, and turning syncing off is an event, both for the site
	syncErrors, err := controlDB.Logs.Queries.SearchErrors(ctx, logsdb.SearchErrorsParams{SiteID: &siteID, Limit: 10})
	require.NoError(t, err)
	require.Len(t, syncErrors, 1)
	require.Equal(t, `{"kind":"connector","connector_type":"dexcom"}`, *syncErrors[0].Attributes)
	events, err := controlDB.Logs.Queries.SearchEvents(ctx, logsdb.SearchEventsParams{SiteID: &siteID, Limit: 10})
	require.NoError(t, err)
	require.Equal(t, "connector_turned_off", events[0].Name)

	// Saving working details clears the error, and the next sync is a first sync again
	require.NoError(t, q.UpdateConnector(ctx, controldb.UpdateConnectorParams{SiteID: siteID, Type: typ, UrlOrUsernameEnc: enc(dexcom_client.TestValidUsername), Enabled: true}))
	s = get()

	// It has failed before, so the retry waits out the doubled backoff rather than running at once
	require.Equal(t, int64(120), s.SyncIntervalSeconds)
	require.False(t, due())
	require.Nil(t, s.FirstSuccessfulSyncTime)
	require.Empty(t, s.PreviousSyncErrors)
	require.NoError(t, runSync())
	s = get()
	require.NotNil(t, s.FirstSuccessfulSyncTime)
	require.True(t, s.Enabled)

	// After the first sync has worked, a failure only backs off, it does not turn syncing off
	require.NoError(t, q.UpdateConnectorReportsFailure(ctx, controldb.UpdateConnectorReportsFailureParams{SiteID: siteID, Type: typ, ErrorMessages: errors.ErrorList{"remote is down"}}))
	s = get()
	require.True(t, s.Enabled)
	require.Equal(t, int64(120), s.SyncIntervalSeconds)

	// Deleting clears it, so a deleted source does not keep its first successful sync time
	require.NoError(t, q.DeleteConnector(ctx, controldb.DeleteConnectorParams{SiteID: siteID, Type: typ}))
	require.Nil(t, get().FirstSuccessfulSyncTime)
}

// Dexcom refusing a source's username or password is its owner's to fix on the Connectors page, which says why, so it is a
// warning and not an error, which would go to Sentry
func TestARefusedDexcomLoginIsAWarning(t *testing.T) {
	cfg, controlDB, site := testutils.NewControlDBWithSite(t)
	q, ctx, siteID, typ := controlDB.Queries, t.Context(), site.SiteDB.Subdomain, consts.ConnectorTypeDexcom
	jobs := connectors.SetupConnectorJobManager(cfg, controlDB)
	enc := func(s string) *string {
		v, err := cfg.Security.Encrypt(s)
		require.NoError(t, err)
		return &v
	}
	var buf bytes.Buffer
	before := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&buf, nil)))
	t.Cleanup(func() { slog.SetDefault(before) })

	require.NoError(t, q.CreateConnector(ctx, controldb.CreateConnectorParams{SiteID: siteID, Type: typ, UrlOrUsernameEnc: enc("bad-user"), PasswordEnc: enc("pass"), Enabled: true}))
	require.Error(t, jobs.RunConnector(ctx, connectors.ConnectorJobParams{SiteID: siteID, Type: typ}))
	require.Contains(t, buf.String(), `level=WARN msg="error syncing user data"`)
	require.NotContains(t, buf.String(), "level=ERROR")
}

// A production server never syncs the virtual source's made-up glucose, though a site has one saved, as a
// development, test or staging server does
func TestVirtualSourceIsNotSyncedInProduction(t *testing.T) {
	for _, mode := range []consts.ServerMode{consts.Production, consts.Test} {
		t.Run(string(mode), func(t *testing.T) {
			testCfg, controlDB, site := testutils.NewControlDBWithSite(t)
			cfg := *testCfg
			cfg.Mode = mode
			q, ctx, siteID := controlDB.Queries, t.Context(), site.SiteDB.Subdomain
			virtual := consts.ConnectorTypeVirtual
			require.NoError(t, q.CreateConnector(ctx, controldb.CreateConnectorParams{SiteID: siteID, Type: virtual,
				UrlOrUsernameEnc: new("virtual"), PasswordEnc: new("virtual"), Enabled: true}))

			require.NoError(t, connectors.SetupConnectorJobManager(&cfg, controlDB).RunConnector(ctx, connectors.ConnectorJobParams{SiteID: siteID, Type: virtual}))
			source, err := q.GetConnector(ctx, controldb.GetConnectorParams{SiteID: siteID, Type: virtual})
			require.NoError(t, err)
			require.Equal(t, mode != consts.Production, source.FirstSuccessfulSyncTime != nil)
		})
	}
}
