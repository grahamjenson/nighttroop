package sync_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	ntsync "nighttroop.t1d.nz/background/sync"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/testutils"
)

func TestVirtualDocumentsAreRepeatable(t *testing.T) {
	end := time.Date(2026, 1, 2, 3, 1, 5, 0, time.UTC)
	docs := ntsync.VirtualDocuments(end, 30*time.Minute)
	// 4 window documents plus an entry and a device status for 6 slots.
	require.Len(t, docs, 4+2*6)
	require.Equal(t, docs, ntsync.VirtualDocuments(end.Add(2*time.Minute), 30*time.Minute))
}

func TestSyncVirtualAddsOneSlotAndDedupes(t *testing.T) {
	config, controlDB, siteInformation, err := testutils.GetConfigControlDBAndSiteDBForTest()
	require.NoError(t, err)
	siteDB, siteModel, err := controlDB.GetSite(siteInformation.SiteDB.Subdomain)
	require.NoError(t, err)
	manager := ntsync.SetupSyncJobManager(config, controlDB)
	syncInfo := ntsync.NewSyncJobInfo(siteModel.ID, consts.SyncSourceTypeVirtual, siteModel, siteDB, controlDB)

	for range 2 {
		require.NoError(t, manager.SyncUserData(syncInfo))
	}

	device := ntsync.VirtualDevice
	readings, err := siteDB.Queries.SearchGlucoseReadings(context.TODO(), sitedb.SearchGlucoseReadingsParams{DeviceEq: &device, Limit: 10})
	require.NoError(t, err)
	require.Len(t, readings, 1)
	reports, err := siteDB.Queries.SearchDeviceReports(context.TODO(), sitedb.SearchDeviceReportsParams{DeviceEq: &device, Limit: 10})
	require.NoError(t, err)
	require.Len(t, reports, 1)
	groups, err := siteDB.Queries.SearchEventGroups(context.TODO(), sitedb.SearchEventGroupsParams{DeviceEq: &device, Limit: 10})
	require.NoError(t, err)
	require.Len(t, groups, 3)
	profiles, err := siteDB.Queries.SearchProfileStores(context.TODO(), sitedb.SearchProfileStoresParams{DeviceEq: &device, Limit: 10})
	require.NoError(t, err)
	require.Len(t, profiles, 0)
}
