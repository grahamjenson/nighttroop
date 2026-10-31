package sync_test

import (
	"bytes"
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/background/sync"
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
	require.NoError(t, q.CreateSyncSource(ctx, controldb.CreateSyncSourceParams{SiteID: siteID, Type: consts.SyncSourceTypeVirtual, Enabled: true}))

	require.NoError(t, sync.SetupSyncJobManager(cfg, controlDB).SyncSource(ctx, sync.SyncJobParams{SiteID: siteID, Type: consts.SyncSourceTypeVirtual}))
	require.Regexp(t, `msg="finished sync".* job_id=[0-9a-f]{16} job=sync site_id=`+siteID+` sync_type=virtual`, logs.String())
}
