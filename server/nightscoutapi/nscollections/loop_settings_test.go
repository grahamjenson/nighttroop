package nscollections

import (
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/models/nightscoutmodels"
	"nighttroop.t1d.nz/utils/testutils"
	"nighttroop.t1d.nz/utils/types"
)

// An app's settings are kept only when they change from those in force when they were sent, each app on its own, so
// the settings in force at a time are its last change at or before it
func TestRecordLoopSettings(t *testing.T) {
	_, _, site := testutils.NewControlDBWithSite(t)
	q, ctx := site.SiteDB.Queries(), t.Context()
	settings := func(maxBolus float64) *nightscoutmodels.LoopSettings {
		return &nightscoutmodels.LoopSettings{MaximumBolus: ptr(types.FlexibleFloat64(maxBolus))}
	}
	inForce := func(app string, at int64) *sitedb.LoopSetting {
		row, err := q.GetLoopSettingsAt(ctx, sitedb.GetLoopSettingsAtParams{App: app, At: at})
		require.NoError(t, err)
		return row
	}
	// record gives the row the settings sent are, which is the one in force at their time
	record := func(app string, s any, at int64) {
		id, err := recordLoopSettings(q, ctx, app, fromProfile, s, at, ptr("phone"))
		require.NoError(t, err)
		require.Equal(t, inForce(app, at).ID, *id)
	}

	for _, step := range []struct {
		name     string
		app      string
		settings any
		at       int64
		// inForceFrom is when the settings in force at the step's time were first sent, once it is recorded
		inForceFrom int64
	}{
		{"the first settings are kept", loopApp, settings(5), 1000, 1000},
		{"the same settings again are not", loopApp, settings(5), 2000, 1000},
		{"changed settings are", loopApp, settings(6), 3000, 3000},
		{"changed back are too", loopApp, settings(5), 4000, 4000},
		{"settings from before the first, as from a backfill, are kept when they differ", loopApp, settings(4), 500, 500},
		{"another app's are its own", trioApp, &nightscoutmodels.RemoteSettings{TeamID: ptr(types.ShortString("team"))}, 2000, 2000},
	} {
		t.Run(step.name, func(t *testing.T) {
			record(step.app, step.settings, step.at)
			require.Equal(t, step.inForceFrom, inForce(step.app, step.at).OccurredAt)
		})
	}

	// Loop's settings in force at each time are its last change at or before it
	require.JSONEq(t, `{"maximumBolus":4}`, inForce(loopApp, 999).SettingsJson)
	require.JSONEq(t, `{"maximumBolus":5}`, inForce(loopApp, 2500).SettingsJson)
	require.JSONEq(t, `{"maximumBolus":6}`, inForce(loopApp, 3500).SettingsJson)
	require.Equal(t, "profile", inForce(loopApp, 3500).Source)
	require.Equal(t, ptr("phone"), inForce(loopApp, 3500).CreatedByDeviceName)
	_, err := q.GetLoopSettingsAt(ctx, sitedb.GetLoopSettingsAtParams{App: loopApp, At: 100})
	require.Error(t, err, "nothing is in force before the first")

	// Settings too large to keep have no row
	id, err := recordLoopSettings(q, ctx, aapsApp, fromDeviceStatus, map[string]string{"big": string(make([]byte, loopSettingsLimit))}, 1000, nil)
	require.NoError(t, err)
	require.Nil(t, id)
}
