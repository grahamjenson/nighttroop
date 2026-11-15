package sitedb_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/utils/testutils"
)

// A value that means none is stored as NULL, as one not sent is: a CGM's noise or rssi of 0, a heart rate of 0, and an
// empty string. An update with one leaves the value stored as it is. A filtered or unfiltered of 0, as AAPS sends for
// every reading, is kept as sent.
func TestValuesMeaningNoneAreNotStored(t *testing.T) {
	_, _, site := testutils.NewControlDBWithSite(t)
	q, ctx, now := site.SiteDB.Queries(), context.Background(), time.Now().UnixMilli()
	zero, noise, empty, serial, heartRate, by := 0.0, 1.0, "", "1234", int64(0), "test"

	reading, err := q.CreateGlucoseReading(ctx, sitedb.CreateGlucoseReadingParams{ID: "r", Mgdl: 100, ReadingType: "sgv",
		Unfiltered: &zero, Filtered: &zero, Rssi: &zero, Noise: &noise, OccurredAt: now, CreatedByIdentityName: "test"})
	require.NoError(t, err)
	require.Equal(t, &zero, reading.Unfiltered)
	require.Equal(t, &zero, reading.Filtered)
	require.Nil(t, reading.Rssi)
	require.Equal(t, &noise, reading.Noise)

	reading, err = q.UpdateGlucoseReading(ctx, sitedb.UpdateGlucoseReadingParams{ID: "r", Noise: &zero, UpdatedByIdentityName: &by})
	require.NoError(t, err)
	require.Equal(t, &noise, reading.Noise)

	group, err := q.CreateEventGroup(ctx, sitedb.CreateEventGroupParams{ID: "g", GroupType: "Temp Basal",
		PumpID: &empty, PumpSerial: &serial, OccurredAt: now, CreatedByIdentityName: "test"})
	require.NoError(t, err)
	require.Nil(t, group.PumpID)
	require.Equal(t, &serial, group.PumpSerial)

	// A treatment rewritten with its merged document has none of what it was rewritten without
	group, err = q.RewriteEventGroup(ctx, sitedb.RewriteEventGroupParams{ID: "g", GroupType: "Temp Basal", PumpSerial: &empty,
		OccurredAt: now, CreatedByIdentityName: by})
	require.NoError(t, err)
	require.Nil(t, group.PumpSerial)

	activity, err := q.CreateActivity(ctx, sitedb.CreateActivityParams{ID: "a", ActivityType: "exercise", HeartRate: &heartRate,
		OccurredAt: now, CreatedByIdentityName: "test"})
	require.NoError(t, err)
	require.Nil(t, activity.HeartRate)
}
