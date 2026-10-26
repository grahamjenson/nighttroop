package backup_sites_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/background/backup_sites"
	"nighttroop.t1d.nz/data/sql/logsdb"
	"nighttroop.t1d.nz/utils/s3utils"
	"nighttroop.t1d.nz/utils/testutils"
)

func TestBackupSiteBatchRecordsEvent(t *testing.T) {
	cfg, controlDB, site, err := testutils.GetConfigControlDBAndSiteDBForTest()
	require.NoError(t, err)
	ctx, batch := context.TODO(), site.SiteDB.Subdomain[:1]
	jobs := backup_sites.SetupBackupSiteJob(ctx, controlDB, cfg, &s3utils.S3Client{ServerConfig: cfg})

	require.NoError(t, jobs.BackupSiteBatch(batch))

	events, err := controlDB.Logs.Queries.SearchEvents(ctx, logsdb.SearchEventsParams{Limit: 10})
	require.NoError(t, err)
	require.Equal(t, "site_batch_backed_up", events[0].Name)
	require.Nil(t, events[0].SiteID)
	var attributes map[string]any
	require.NoError(t, json.Unmarshal([]byte(*events[0].Attributes), &attributes))
	require.Equal(t, batch, attributes["batch"])
	require.Equal(t, float64(1), attributes["site_count"])
	// S3 is not connected in tests, so the batch was backed up but not uploaded
	require.Equal(t, false, attributes["uploaded"])
	require.Contains(t, attributes, "duration_ms")
}

func TestDeleteUnusedSitesRecordsEvent(t *testing.T) {
	cfg, controlDB, site, err := testutils.GetConfigControlDBAndSiteDBForTest()
	require.NoError(t, err)
	ctx, siteID := context.TODO(), site.SiteDB.Subdomain
	jobs := backup_sites.SetupBackupSiteJob(ctx, controlDB, cfg, &s3utils.S3Client{ServerConfig: cfg})
	// The site has never been used, and counts as unused from the moment it was made
	jobs.DeleteAfterInactivity = -time.Minute

	require.NoError(t, jobs.DeleteUnusedSites(siteID[:1]))

	events, err := controlDB.Logs.Queries.SearchEvents(ctx, logsdb.SearchEventsParams{SiteID: &siteID, Limit: 10})
	require.NoError(t, err)
	require.Equal(t, "site_deleted", events[0].Name)
	require.Equal(t, `{"reason":"unused"}`, *events[0].Attributes)
}
