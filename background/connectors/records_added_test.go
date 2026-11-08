package connectors_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/background/connectors"
	"nighttroop.t1d.nz/clients/dexcom_client"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/data/sql/logsdb"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/testutils"
)

// A sync that adds records records a connector_records_added event with how many, keeps them as its source's last sync's records,
// and adds them to its total; one that adds none records no event
func TestSyncThatAddsRecordsRecordsAnEvent(t *testing.T) {
	cfg, controlDB, site := testutils.NewControlDBWithSite(t)
	q, ctx, siteID := controlDB.Queries, t.Context(), site.SiteDB.Subdomain
	jobs := connectors.SetupConnectorJobManager(cfg, controlDB)
	now := time.Now().UnixMilli()
	require.NoError(t, q.UpdateSiteDateFlags(ctx, controldb.UpdateSiteDateFlagsParams{LastSeenAt: &now, ID: siteID}))
	username, err := cfg.Security.Encrypt(dexcom_client.TestValidUsername)
	require.NoError(t, err)
	password, err := cfg.Security.Encrypt("pass")
	require.NoError(t, err)
	added := func() []string {
		events, err := controlDB.Logs.Queries.SearchEvents(ctx, logsdb.SearchEventsParams{SiteID: &siteID, Search: new("connector_records_added"), Limit: 10})
		require.NoError(t, err)
		attributes := []string{}
		for _, event := range events {
			require.Equal(t, "connector_records_added", event.Name)
			attributes = append(attributes, *event.Attributes)
		}
		return attributes
	}

	// The test Dexcom has no readings, so its sync adds nothing
	require.NoError(t, q.CreateConnector(ctx, controldb.CreateConnectorParams{SiteID: siteID, Type: consts.ConnectorTypeDexcom, UrlOrUsernameEnc: &username, PasswordEnc: &password, Enabled: true}))
	require.NoError(t, jobs.RunConnector(ctx, connectors.ConnectorJobParams{SiteID: siteID, Type: consts.ConnectorTypeDexcom}))
	require.Empty(t, added())
	source := func(typ consts.ConnectorType) *controldb.SiteConnector {
		source, err := q.GetConnector(ctx, controldb.GetConnectorParams{SiteID: siteID, Type: typ})
		require.NoError(t, err)
		return source
	}
	require.Zero(t, source(consts.ConnectorTypeDexcom).TotalRecords)
	require.Equal(t, int64(0), *source(consts.ConnectorTypeDexcom).LastSyncRecords)

	// The virtual connector's first sync fills the last 30 minutes, so it adds records
	require.NoError(t, q.CreateConnector(ctx, controldb.CreateConnectorParams{SiteID: siteID, Type: consts.ConnectorTypeVirtual, Enabled: true}))
	require.NoError(t, jobs.RunConnector(ctx, connectors.ConnectorJobParams{SiteID: siteID, Type: consts.ConnectorTypeVirtual}))
	events := added()
	require.Len(t, events, 1)
	var attributes struct {
		ConnectorType string `json:"connector_type"`
		Records  int    `json:"records"`
	}
	require.NoError(t, json.Unmarshal([]byte(events[0]), &attributes))
	require.Equal(t, "virtual", attributes.ConnectorType)
	require.Positive(t, attributes.Records)
	virtual := source(consts.ConnectorTypeVirtual)
	require.Equal(t, int64(attributes.Records), virtual.TotalRecords)
	require.Equal(t, int64(attributes.Records), *virtual.LastSyncRecords)

	// The next sync's records are its last, and are added to the total
	require.NoError(t, jobs.RunConnector(ctx, connectors.ConnectorJobParams{SiteID: siteID, Type: consts.ConnectorTypeVirtual}))
	next := source(consts.ConnectorTypeVirtual)
	require.Equal(t, virtual.TotalRecords+*next.LastSyncRecords, next.TotalRecords)
}
