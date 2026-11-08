package connectors_test

import (
	"bytes"
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/background/connectors"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/utils/aslog"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/testutils"
)

// A sync's logs say which job ran it, and for which site and source
func TestSyncLogsItsJobAndSite(t *testing.T) {
	cfg, controlDB, site := testutils.NewControlDBWithSite(t)
	q, siteID := controlDB.Queries, site.SiteDB.Subdomain
	var logs bytes.Buffer
	testutils.UseLogHandler(t, aslog.ContextHandler{Handler: slog.NewTextHandler(&logs, nil)})
	ctx := aslog.NewJob(context.Background(), "sync")
	now := time.Now().UnixMilli()
	require.NoError(t, q.UpdateSiteDateFlags(ctx, controldb.UpdateSiteDateFlagsParams{LastSeenAt: &now, ID: siteID}))
	require.NoError(t, q.CreateConnector(ctx, controldb.CreateConnectorParams{SiteID: siteID, Type: consts.ConnectorTypeVirtual, Enabled: true}))

	require.NoError(t, connectors.SetupConnectorJobManager(cfg, controlDB).RunConnector(ctx, connectors.ConnectorJobParams{SiteID: siteID, Type: consts.ConnectorTypeVirtual}))
	require.Regexp(t, `msg="finished sync".* job_id=[0-9a-f]{16} job=sync site_id=`+siteID+` connector_type=virtual`, logs.String())
}
