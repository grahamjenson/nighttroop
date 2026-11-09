package sitedb_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/utils/testutils"
)

// A value that means none is stored as NULL, as one not sent is: a CGM's raw signal of 0, as AAPS sends unfiltered for
// every reading, a heart rate of 0, and an empty string. An update with one leaves the value stored as it is.
func TestValuesMeaningNoneAreNotStored(t *testing.T) {
	_, _, site := testutils.NewControlDBWithSite(t)
	q, ctx, now := site.SiteDB.Queries(), context.Background(), time.Now().UnixMilli()
	zero, noise, empty, serial, heartRate, by := 0.0, 1.0, "", "4241", int64(0), "test"

	reading, err := q.CreateGlucoseReading(ctx, sitedb.CreateGlucoseReadingParams{ID: "r", Mgdl: 100, ReadingType: "sgv",
		Unfiltered: &zero, Filtered: &zero, Rssi: &zero, Noise: &noise, OccurredAt: now, CreatedByIdentityName: "test"})
	require.NoError(t, err)
	require.Nil(t, reading.Unfiltered)
	require.Nil(t, reading.Filtered)
	require.Nil(t, reading.Rssi)
	require.Equal(t, &noise, reading.Noise)

	reading, err = q.UpdateGlucoseReading(ctx, sitedb.UpdateGlucoseReadingParams{ID: "r", Noise: &zero, UpdatedByIdentityName: &by})
	require.NoError(t, err)
	require.Equal(t, &noise, reading.Noise)

	event, err := q.CreateComponentEvent(ctx, sitedb.CreateComponentEventParams{ID: "c", EventType: "Site Change",
		PumpID: &empty, PumpSerial: &serial, OccurredAt: now, CreatedByIdentityName: "test"})
	require.NoError(t, err)
	require.Nil(t, event.PumpID)
	require.Equal(t, &serial, event.PumpSerial)

	event, err = q.UpdateComponentEvent(ctx, sitedb.UpdateComponentEventParams{ID: "c", PumpSerial: &empty, UpdatedByIdentityName: &by})
	require.NoError(t, err)
	require.Equal(t, &serial, event.PumpSerial)

	activity, err := q.CreateActivity(ctx, sitedb.CreateActivityParams{ID: "a", ActivityType: "exercise", HeartRate: &heartRate,
		OccurredAt: now, CreatedByIdentityName: "test"})
	require.NoError(t, err)
	require.Nil(t, activity.HeartRate)
}
