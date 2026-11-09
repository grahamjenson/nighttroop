package nscollections

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/models/nightscoutmodels"
	"nighttroop.t1d.nz/utils/httputils"
	"nighttroop.t1d.nz/utils/testutils"
	"nighttroop.t1d.nz/utils/types"
)

// An app's settings are kept once however often they are sent, by the SHA-256 of the app and the settings, with when
// they were first and last seen
func TestRecordLoopSettings(t *testing.T) {
	_, _, site := testutils.NewControlDBWithSite(t)
	q, ctx := site.SiteDB.Queries(), t.Context()
	settings := func(maxBolus float64) *nightscoutmodels.LoopSettings {
		return &nightscoutmodels.LoopSettings{MaximumBolus: ptr(types.FlexibleFloat64(maxBolus))}
	}
	record := func(app string, s any, at int64) *string {
		id, err := recordLoopSettings(q, ctx, app, fromProfile, s, at, ptr("phone"))
		require.NoError(t, err)
		return id
	}
	get := func(id *string) *sitedb.LoopSetting {
		row, err := q.GetLoopSettings(ctx, sitedb.GetLoopSettingsParams{ID: *id})
		require.NoError(t, err)
		return row
	}

	first := record(loopApp, settings(5), 2000)
	require.Len(t, *first, 64)
	require.Equal(t, loopSettingsID(loopApp, `{"maximumBolus":5}`), *first)

	// The same settings again, later and from a backfill earlier, are the same row, seen over both times
	require.Equal(t, first, record(loopApp, settings(5), 3000))
	require.Equal(t, first, record(loopApp, settings(5), 1000))
	row := get(first)
	require.Equal(t, int64(1000), row.FirstSeenAt)
	require.Equal(t, int64(3000), row.LastSeenAt)
	require.Equal(t, "loop", row.App)
	require.Equal(t, "profile", row.Source)
	require.Equal(t, ptr("phone"), row.CreatedByDeviceName)
	require.JSONEq(t, `{"maximumBolus":5}`, row.SettingsJson)

	// Changed settings are another row, and changing back is the first again
	changed := record(loopApp, settings(6), 4000)
	require.NotEqual(t, first, changed)
	require.Equal(t, first, record(loopApp, settings(5), 5000))
	require.Equal(t, int64(5000), get(first).LastSeenAt)

	// The same settings from another app are its own
	require.NotEqual(t, first, record(trioApp, settings(5), 2000))

	// Empty settings, as AAPS sends in most statuses, none at all, and settings too large to keep have no row
	require.Nil(t, record(aapsApp, map[string]any{}, 1000))
	require.Nil(t, record(loopApp, (*nightscoutmodels.LoopSettings)(nil), 1000))
	require.Nil(t, record(loopApp, &nightscoutmodels.LoopSettings{}, 1000))
	require.Nil(t, record(aapsApp, map[string]string{"big": string(make([]byte, loopSettingsLimit))}, 1000))
	rows, err := q.SearchLoopSettings(ctx, sitedb.SearchLoopSettingsParams{Limit: 10})
	require.NoError(t, err)
	require.Len(t, rows, 3)
}

// A device status is read back with the configuration it refers to, and without it once that is deleted, as the
// report is kept
func TestDeviceStatusConfigurationFromLoopSettings(t *testing.T) {
	_, _, site := testutils.NewControlDBWithSite(t)
	q, ctx := site.SiteDB.Queries(), t.Context()
	siteContext := &httputils.SiteContext{SiteDB: site.SiteDB}
	var status nightscoutmodels.DeviceStatus
	require.NoError(t, json.Unmarshal([]byte(`{"device":"openaps://phone","configuration":{"aps":"SMB","insulin":2}}`), &status))

	report, err := DeviceStatusCollection{}.CreateModel(q, ctx, status, nightscoutmodels.ConverterContext{IdentityName: "aaps"}, nil)
	require.NoError(t, err)
	require.NotNil(t, report.OtherAttributes.ConfigurationID)
	require.JSONEq(t, `{"aps":"SMB","insulin":2}`, string(*DBToNightscoutDeviceStatus(report, siteContext).ConfigurationRaw))

	_, err = q.DeleteLoopSettings(ctx, sitedb.DeleteLoopSettingsParams{ID: *report.OtherAttributes.ConfigurationID})
	require.NoError(t, err)
	report, err = q.GetDeviceReport(ctx, sitedb.GetDeviceReportParams{ID: report.ID})
	require.NoError(t, err)
	require.Nil(t, DBToNightscoutDeviceStatus(report, siteContext).ConfigurationRaw)
}
