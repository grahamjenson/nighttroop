package db_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/events"
	"nighttroop.t1d.nz/data/sql/livedb"
	"nighttroop.t1d.nz/data/sql/logsdb"
	"nighttroop.t1d.nz/utils/testutils"
)

// A resource going over its alert level is an event, as is it coming back under, but not staying over; the banner's
// high resources are the last check's, and checks older than a few days are forgotten
func TestRecordResources(t *testing.T) {
	_, controlDB := testutils.NewControlDB(t)
	cfg := *controlDB.ServerConfig
	cfg.MemoryAlertPercent, cfg.DiskAlertPercent, cfg.CPUPeggedAlertSeconds, cfg.DiskAlertFreeGB = 60, 60, 20, 100
	controlDB.ServerConfig = &cfg
	ctx := context.Background()
	start := time.Now().Add(-4 * 24 * time.Hour)

	percent := func(p float64) *float64 { return &p }
	const gb = 1 << 30
	checks := []struct {
		memory, disk *float64
		diskTotal    *int64
		pegged       int64
		want         []string
		high         []string
	}{
		{memory: percent(40), disk: percent(30), pegged: 0},
		{memory: percent(70), disk: percent(30), pegged: 0, want: []string{"resource_high memory"}, high: []string{"memory"}},
		{memory: percent(75), disk: percent(30), pegged: 0, high: []string{"memory"}},
		{memory: percent(50), disk: percent(61), pegged: 25, want: []string{"resource_high disk", "resource_high cpu", "resource_recovered memory"}, high: []string{"disk", "cpu"}},
		// Memory that cannot be read, as on a Mac, is neither high nor recovered
		{disk: percent(30), pegged: 5, want: []string{"resource_recovered disk", "resource_recovered cpu"}},
		// A disk 87% used with 124 GB still free does not warn, but with 12 GB free it does
		{disk: percent(87), diskTotal: new(int64(954 * gb))},
		{disk: percent(87), diskTotal: new(int64(95 * gb)), want: []string{"resource_high disk"}, high: []string{"disk"}},
	}
	seen := 0
	for i, check := range checks {
		at := start.Add(time.Duration(i) * 24 * time.Hour)
		require.NoError(t, controlDB.RecordResources(ctx, livedb.CreateResourceReadingParams{At: at.UnixMilli(), CPUPercent: 50,
			CPUPeggedSeconds: check.pegged, MemoryPercent: check.memory, DiskPercent: check.disk, DiskTotalBytes: check.diskTotal}))

		recorded, err := controlDB.Logs.Queries.SearchEvents(ctx, logsdb.SearchEventsParams{Limit: 100})
		require.NoError(t, err)
		got := []string{}
		for _, event := range recorded {
			if event.Name == string(events.ResourceHigh) || event.Name == string(events.ResourceRecovered) {
				got = append(got, event.Name)
			}
		}
		require.Len(t, got, seen+len(check.want), "check %d", i)
		seen = len(got)

		high := []string{}
		for _, level := range controlDB.HighResources(ctx) {
			high = append(high, level.Resource)
		}
		require.ElementsMatch(t, check.high, high, "check %d", i)
	}

	// The newest events say which resource, how high and of what total
	recovered, high := string(events.ResourceRecovered), string(events.ResourceHigh)
	recorded, err := controlDB.Logs.Queries.SearchEvents(ctx, logsdb.SearchEventsParams{Limit: 1, Type: &recovered})
	require.NoError(t, err)
	require.Equal(t, `{"resource":"cpu","value":5,"limit":20}`, *recorded[0].Attributes)
	recorded, err = controlDB.Logs.Queries.SearchEvents(ctx, logsdb.SearchEventsParams{Limit: 1, Type: &high})
	require.NoError(t, err)
	require.Equal(t, `{"resource":"disk","value":87,"limit":60,"total":102005473280}`, *recorded[0].Attributes)

	// Seven checks a day apart, so the first three are over 3 days older than the newest, and forgotten
	kept, err := controlDB.Live.Queries.ResourceReadingsSince(ctx, livedb.ResourceReadingsSinceParams{Since: 0})
	require.NoError(t, err)
	require.Len(t, kept, 4)
	require.Equal(t, start.Add(3*24*time.Hour).UnixMilli(), kept[0].At)
}
