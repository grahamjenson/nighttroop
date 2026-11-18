package connectors

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/clients/nightscout_client"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/security"
	"nighttroop.t1d.nz/utils/testutils"
)

// A usual sync that works sets when the next entry is due, 5 minutes after the entries last changed; a timed sync only
// clears it once it sees a new entry; a failed sync, or a Nightscout that does not say, clears it
func TestTimeNextEntry(t *testing.T) {
	const modified, due = int64(1_000_000), int64(1_000_000 + 300_000)
	tests := []struct {
		name     string
		before   int64
		modified int64
		timed    bool
		err      error
		after    int64
	}{
		{"a usual sync sets it", 0, modified, false, nil, due},
		{"a usual sync moves it to the newer entry", due, modified + 300_000, false, nil, due + 300_000},
		{"a timed sync before the entry keeps it", due, modified, true, nil, due},
		{"a timed sync that sees the entry clears it", due, modified + 300_000, true, nil, 0},
		{"a timed sync does not set it", 0, modified, true, nil, 0},
		{"a usual sync that fails clears it", due, modified, false, errors.New("HTTP 500"), 0},
		{"a timed sync that fails clears it", due, modified, true, errors.New("timeout"), 0},
		{"no entries time clears it", due, 0, false, nil, 0},
	}
	for _, test := range tests {
		c := &nightscoutClient{}
		c.nextEntryAt.Store(test.before)
		c.timeNextEntry(test.modified, test.timed, test.err)
		require.Equal(t, test.after, c.nextEntryAt.Load(), test.name)
	}
}

// A site is due a timed sync from 15 seconds before its next entry is due to 90 seconds after, at most every 10 seconds
func TestEntriesDue(t *testing.T) {
	cfg, controlDB, _ := testutils.NewControlDBWithSite(t)
	jobs := SetupConnectorJobManager(cfg, controlDB)
	due := time.UnixMilli(2_000_000_000_000)
	c := jobs.nightscoutClientFor(&ConnectorJobInfo{SiteID: "timedsite", urlOrUsername: nightscout_client.TestValidURL, password: "token"})
	jobs.nightscoutClientFor(&ConnectorJobInfo{SiteID: "untimeds", urlOrUsername: nightscout_client.TestValidURL, password: "token"})
	c.nextEntryAt.Store(due.UnixMilli())

	require.Empty(t, jobs.entriesDue(due.Add(-16*time.Second)), "too early")
	require.Equal(t, []string{"timedsite"}, jobs.entriesDue(due.Add(-15*time.Second)))
	require.Empty(t, jobs.entriesDue(due.Add(-10*time.Second)), "within 10 seconds of the last")
	require.Equal(t, []string{"timedsite"}, jobs.entriesDue(due.Add(-5*time.Second)))
	require.Equal(t, []string{"timedsite"}, jobs.entriesDue(due.Add(90*time.Second)))
	require.Empty(t, jobs.entriesDue(due.Add(101*time.Second)), "too late")
}

// A site whose client is syncing now, or whose last sync was slow, gets no timed sync until that changes
func TestEntriesDueSkipsABusyOrSlowNightscout(t *testing.T) {
	cfg, controlDB, _ := testutils.NewControlDBWithSite(t)
	jobs := SetupConnectorJobManager(cfg, controlDB)
	due := time.UnixMilli(2_000_000_000_000)
	timed := func(siteID string) *nightscoutClient {
		c := jobs.nightscoutClientFor(&ConnectorJobInfo{SiteID: siteID, urlOrUsername: "https://" + siteID + ".example.com", password: "token"})
		c.nextEntryAt.Store(due.UnixMilli())
		return c
	}
	busy, slow := timed("busysite"), timed("slowsite")
	busy.syncing.Store(true)
	slow.lastSyncTook.Store(slowSync.Milliseconds() + 1)
	require.Empty(t, jobs.entriesDue(due))
	busy.syncing.Store(false)
	slow.lastSyncTook.Store(slowSync.Milliseconds())
	require.ElementsMatch(t, []string{"busysite", "slowsite"}, jobs.entriesDue(due))
}

// Every sync, usual or timed, of every connector is limited by CONNECTOR_SYNC_PER_MINUTE: a Nightscout host however
// many sites sync from it, and however it is written, or else each connector
func TestConnectorSyncsAreRateLimited(t *testing.T) {
	shared, controlDB, site := testutils.NewControlDBWithSite(t)
	// Its own limiters, so the other tests' syncs of the fake Nightscout are not limited by it
	cfg := *shared
	cfg.RateLimiters = config.NewRateLimiters(cfg)
	cfg.RateLimiters.SetRates(map[string]config.Rate{"connector_sync": {PerMinute: 6, Burst: 1}})
	q, ctx, typ := controlDB.Queries, t.Context(), consts.ConnectorTypeNightscout
	jobs := SetupConnectorJobManager(&cfg, controlDB)
	other, err := controlDB.CreateSite(ctx)
	require.NoError(t, err)
	attempted := func(siteID string) *int64 {
		connector, err := q.GetConnector(ctx, controldb.GetConnectorParams{SiteID: siteID, Type: typ})
		require.NoError(t, err)
		return connector.LastAttemptedSync
	}
	for _, siteID := range []string{site.SiteDB.Subdomain, other.SiteDB.Subdomain} {
		url, err := cfg.Security.EncryptFor(security.PurposeConnector, siteID, nightscout_client.TestValidURL)
		require.NoError(t, err)
		token, err := cfg.Security.EncryptFor(security.PurposeConnector, siteID, "token")
		require.NoError(t, err)
		require.NoError(t, q.CreateConnector(ctx, controldb.CreateConnectorParams{SiteID: siteID, Type: typ, UrlOrUsernameEnc: &url, PasswordEnc: &token, Enabled: true}))
	}

	require.NoError(t, jobs.RunConnector(ctx, ConnectorJobParams{SiteID: site.SiteDB.Subdomain, Type: typ}))
	synced := attempted(site.SiteDB.Subdomain)
	require.NotNil(t, synced)
	require.NoError(t, jobs.RunConnector(ctx, ConnectorJobParams{SiteID: other.SiteDB.Subdomain, Type: typ}))
	require.Nil(t, attempted(other.SiteDB.Subdomain), "another site on the same host waits")
	require.NoError(t, jobs.RunConnector(ctx, ConnectorJobParams{SiteID: site.SiteDB.Subdomain, Type: typ, Timed: true}))
	require.Equal(t, *synced, *attempted(site.SiteDB.Subdomain), "a timed sync waits too")

	nightscout := &controldb.SiteConnector{Type: typ, SiteID: "aaaaaaaa"}
	require.Equal(t, syncLimitKey(nightscout, "https://ns.example.com/"), syncLimitKey(&controldb.SiteConnector{Type: typ, SiteID: "bbbbbbbb"}, "https://NS.example.com./api/v1"))
	require.NotEqual(t, syncLimitKey(&controldb.SiteConnector{Type: consts.ConnectorTypeDexcom, SiteID: "aaaaaaaa"}, "user"), syncLimitKey(&controldb.SiteConnector{Type: consts.ConnectorTypeDexcom, SiteID: "bbbbbbbb"}, "user"), "Dexcom is limited by connector")
}

// A sync on the usual schedule times the next entry; a timed sync runs only for an enabled connector with no failed
// syncs in a row, so a failing or disabled Nightscout is not synced more often
func TestTimedSyncOnlyForAWorkingConnector(t *testing.T) {
	cfg, controlDB, site := testutils.NewControlDBWithSite(t)
	q, ctx, siteID, typ := controlDB.Queries, t.Context(), site.SiteDB.Subdomain, consts.ConnectorTypeNightscout
	jobs := SetupConnectorJobManager(cfg, controlDB)
	enc := func(s string) *string {
		v, err := cfg.Security.EncryptFor(security.PurposeConnector, siteID, s)
		require.NoError(t, err)
		return &v
	}
	attempted := func() *int64 {
		connector, err := q.GetConnector(ctx, controldb.GetConnectorParams{SiteID: siteID, Type: typ})
		require.NoError(t, err)
		return connector.LastAttemptedSync
	}
	require.NoError(t, q.CreateConnector(ctx, controldb.CreateConnectorParams{SiteID: siteID, Type: typ, UrlOrUsernameEnc: enc(nightscout_client.TestValidURL), PasswordEnc: enc("token"), Enabled: true}))

	require.NoError(t, jobs.RunConnector(ctx, ConnectorJobParams{SiteID: siteID, Type: typ}))
	c, _ := jobs.nightscoutClients.Get(siteID)
	require.Equal(t, nightscout_client.TestLastModified+entryEvery.Milliseconds(), c.nextEntryAt.Load())

	// A timed sync of a working connector runs
	before := attempted()
	time.Sleep(2 * time.Millisecond)
	require.NoError(t, jobs.RunConnector(ctx, ConnectorJobParams{SiteID: siteID, Type: typ, Timed: true}))
	require.NotEqual(t, *before, *attempted())

	// One with a failed sync, or disabled, does not
	ms := int64(10)
	require.NoError(t, q.UpdateConnectorReportsFailure(ctx, controldb.UpdateConnectorReportsFailureParams{ErrorMessages: []string{"HTTP 500"}, ElapsedMs: &ms, SiteID: siteID, Type: typ}))
	before = attempted()
	time.Sleep(2 * time.Millisecond)
	require.NoError(t, jobs.RunConnector(ctx, ConnectorJobParams{SiteID: siteID, Type: typ, Timed: true}))
	require.Equal(t, *before, *attempted(), "failing")
	require.NoError(t, q.UpdateConnector(ctx, controldb.UpdateConnectorParams{SiteID: siteID, Type: typ, PasswordEnc: enc("token"), Enabled: false}))
	require.NoError(t, jobs.RunConnector(ctx, ConnectorJobParams{SiteID: siteID, Type: typ, Timed: true}))
	require.Equal(t, *before, *attempted(), "disabled")
}
