package backup_sites_test

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/background/backup_sites"
	"nighttroop.t1d.nz/data/sql/logsdb"
	"nighttroop.t1d.nz/utils/s3utils"
	"nighttroop.t1d.nz/utils/testutils"
)

func TestBackupSiteBatchRecordsEvent(t *testing.T) {
	cfg, controlDB, site := testutils.NewControlDBWithSite(t)
	ctx, batch := t.Context(), site.SiteDB.Subdomain[:1]
	jobs := backup_sites.SetupBackupSiteJob(ctx, controlDB, cfg, &s3utils.S3Client{ServerConfig: cfg})
	var logs bytes.Buffer
	testutils.UseLogHandler(t, slog.NewTextHandler(&logs, nil))

	started := time.Now().UnixMilli()
	require.NoError(t, jobs.BackupSiteBatch(ctx, batch))

	// One log lists the sites backed up, and as nothing went wrong, nothing else is logged
	lines := strings.Split(strings.TrimSpace(logs.String()), "\n")
	require.Len(t, lines, 1, logs.String())
	require.Contains(t, lines[0], `msg="Backing up site DBs" sites=[`+site.SiteDB.Subdomain+`]`)

	events, err := controlDB.Logs.Queries.SearchEvents(ctx, logsdb.SearchEventsParams{Limit: 10})
	require.NoError(t, err)
	require.Equal(t, "site_batch_backed_up", events[0].Name)
	require.Nil(t, events[0].SiteID)
	require.Equal(t, batch, *events[0].Batch, "so each site in the batch can see it")
	var attributes map[string]any
	require.NoError(t, json.Unmarshal([]byte(*events[0].Attributes), &attributes))
	require.Equal(t, batch, attributes["batch"])
	require.Equal(t, float64(1), attributes["site_count"])
	// S3 is not connected in tests, so the batch was backed up but not uploaded
	require.Equal(t, false, attributes["uploaded"])
	// It started when the backup did, and finished when it was recorded
	require.GreaterOrEqual(t, events[0].StartedAt, started)
	require.NotNil(t, events[0].FinishedAt)
	require.GreaterOrEqual(t, *events[0].FinishedAt, events[0].StartedAt)
}

func TestDeleteUnusedSitesRecordsEvent(t *testing.T) {
	cfg, controlDB, site := testutils.NewControlDBWithSite(t)
	ctx, siteID := t.Context(), site.SiteDB.Subdomain
	jobs := backup_sites.SetupBackupSiteJob(ctx, controlDB, cfg, &s3utils.S3Client{ServerConfig: cfg})
	// The site has never been used, and counts as unused from the moment it was made
	jobs.DeleteAfterInactivity = -time.Minute

	require.NoError(t, jobs.DeleteUnusedSites(ctx, siteID[:1]))

	events, err := controlDB.Logs.Queries.SearchEvents(ctx, logsdb.SearchEventsParams{SiteID: &siteID, Limit: 10})
	require.NoError(t, err)
	require.Equal(t, "site_deleted", events[0].Name)
	require.Equal(t, `{"reason":"unused"}`, *events[0].Attributes)
}
