package nsv2

import (
	"cmp"
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/data/sql/sqltiming"
	"nighttroop.t1d.nz/utils/testutils"
)

// openAPSStatus is an AAPS device status's openaps, as it sends one every 5 minutes, with its iob, suggested and enacted
func openAPSStatus(at time.Time) string {
	stamp := at.UTC().Format("2006-01-02T15:04:05.000Z")
	run := `"algorithm": "SMB", "timestamp": "` + stamp + `", "bg": 150, "tick": "-6", "eventualBG": 41, "targetBG": 123,
		"insulinReq": 0, "deliverAt": "` + stamp + `", "sensitivityRatio": 1.2055221584632745, "duration": 120, "rate": 0,
		"reason": "Dosing sensitivity: 330.0 using eventual BG;COB: 0, Dev: -0.6, BGI: -0.3, ISF: 7.5, CR: 11.5, Target: 6.8",
		"predBGs": {
			"IOB": [150, 144, 138, 132, 126, 120, 115, 109, 104, 99, 94, 89, 85, 80, 76, 71, 67, 62, 58, 53, 49, 44, 40, 39],
			"ZT": [150, 145, 141, 136, 131, 127, 123, 118, 114, 111, 107, 104, 101, 98, 95, 93, 92, 90, 89, 89, 89, 89, 90, 91, 92, 94, 96, 98, 101, 103, 106, 109, 112, 116, 119, 123],
			"UAM": [150, 144, 138, 132, 126, 120, 115, 109, 104, 99, 94, 89, 85, 80, 76, 71, 67, 62, 58, 53, 49, 44, 40, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39, 39]
		},
		"COB": 0, "IOB": 0.736, "variable_sens": 134.4,
		"consoleLog": ["Autosens ratio: 1.2055221584632745; ", "Adjusting basal from 0.35 to 0.42193275546214604; ", "target_bg from 135.9 to 123.0; "],
		"consoleError": ["SMB enabled due to enableSMB_always", "profile.sens: 134.4, sens: 134.4, CSF: 11.68695652173913", "minGuardBG -1.1 projected below 4.5 - disabling SMB"]`
	return `{"iob": {"iob": 0.736, "basaliob": -0.332, "activity": 0.007, "time": "` + stamp + `"},
		"suggested": {` + run + `}, "enacted": {` + run + `, "received": true}}`
}

// seedProperties gives an AAPS site what every property is made from: readings, change treatments and a loop's status
func seedProperties(t testing.TB) (context.Context, *sitedb.Queries) {
	_, _, site := testutils.NewControlDBWithSite(t)
	ctx, q, now := t.Context(), site.SiteDB.Queries(), time.Now()
	for i := range int64(12) {
		_, err := q.CreateGlucoseReading(ctx, sitedb.CreateGlucoseReadingParams{ID: fmt.Sprint("r", i), Mgdl: float64(120 + i),
			ReadingType: "sgv", Trend: new(int64(4)), OccurredAt: now.Add(-time.Duration(i) * 5 * time.Minute).UnixMilli(), CreatedByIdentityName: "test"})
		require.NoError(t, err)
	}
	for i, groupType := range []string{"Site Change", "Insulin Change", "Sensor Change", "Sensor Start"} {
		_, err := q.CreateEventGroup(ctx, sitedb.CreateEventGroupParams{ID: fmt.Sprint("g", i), GroupType: groupType,
			OccurredAt: now.Add(-time.Duration(i+1) * 24 * time.Hour).UnixMilli(), CreatedByIdentityName: "test"})
		require.NoError(t, err)
	}
	_, err := q.CreateDeviceReport(ctx, sitedb.CreateDeviceReportParams{ID: "d", LoopName: new("openaps"),
		LoopStatusJson: new(openAPSStatus(now.Add(-time.Minute))), LoopIob: new(0.736), LoopCob: new(0.0),
		DeviceBatteryPercent: new(88.0), PumpJson: new(`{"clock": "` + now.UTC().Format(time.RFC3339) + `", "reservoir": 38, "status": {"status": "normal"}}`),
		OccurredAt: now.Add(-time.Minute).UnixMilli(), CreatedByDeviceName: new("openaps://Test Phone"), CreatedByIdentityName: "test"})
	require.NoError(t, err)
	return ctx, q
}

// BenchmarkProperties is a watch face or Nightguard polling an AAPS site's properties, a client asking for some of
// them, as LoopFollow asks for dbsize, and only the openaps property.
// Run with go test -run XXX -bench Properties -benchmem ./server/nightscoutapi/nsapi/nsv2
func BenchmarkProperties(b *testing.B) {
	ctx, q := seedProperties(b)
	properties := Properties(ctx, q, time.Now())
	require.Equal(b, "Enacted", properties["openaps"].(map[string]any)["status"].(LoopState).Label)
	require.NotNil(b, properties["openaps"].(map[string]any)["lastPredBGs"])
	for _, names := range [][]string{nil, {"dbsize"}, {"cage"}, {"bgnow", "delta", "direction"}, {"iob", "cob"}} {
		b.Run(cmp.Or(strings.Join(names, ","), "all"), func(b *testing.B) {
			for b.Loop() {
				Properties(ctx, q, time.Now(), names...)
			}
		})
	}
	dr := LastDeviceStatusFromPWD(q, ctx)
	b.Run("openaps", func(b *testing.B) {
		for b.Loop() {
			GetOpenAPS(dr, time.Now())
		}
	})
}

// A request for some properties gives what the list of all has for them, and reads only what they are made from
func TestPropertiesAskedFor(t *testing.T) {
	ctx, q := seedProperties(t)
	sqltiming.SetSamplePercent(100)
	t.Cleanup(func() { sqltiming.SetSamplePercent(0) })
	queries := func() (count int64) {
		for _, query := range sqltiming.Slowest(math.MaxInt) {
			count += query.Count
		}
		return count
	}
	now, before := time.Now(), queries()
	all := Properties(ctx, q, now)
	require.Equal(t, int64(8), queries()-before)
	require.Len(t, all, 13)
	for names, reads := range map[string]int64{"dbsize": 0, "cage": 1, "sage": 2, "upbat": 1, "delta": 1,
		"bgnow,delta,direction": 1, "iob,cob,pump,openaps,loop": 1, "cage,notaproperty,bgnow": 2, ",": 0} {
		want := map[string]any{}
		for _, name := range strings.Split(names, ",") {
			if property, ok := all[name]; ok {
				want[name] = property
			}
		}
		before := queries()
		require.Equal(t, want, Properties(ctx, q, now, strings.Split(names, ",")...), names)
		require.Equal(t, reads, queries()-before, names)
	}
}

// Each run of a loop status has its own moment, and the status is shown as it was sent
func TestLoopRunMoments(t *testing.T) {
	now := time.Date(2026, 11, 23, 6, 0, 0, 0, time.UTC)
	asJSON := func(v any) string { raw, err := json.Marshal(v); require.NoError(t, err); return string(raw) }
	openAPS := GetOpenAPS(&sitedb.DeviceReport{LoopName: new("openaps"), LoopStatusJson: new(`{
		"iob": {"iob": 1.2, "time": "2026-11-23T05:58:00.000Z"},
		"suggested": {"timestamp": "2026-11-23T05:57:00.000Z", "eventualBG": 100, "predBGs": {"IOB": [100, 99]}},
		"enacted": {"timestamp": "2026-11-23T05:58:00.000Z", "eventualBG": 110, "received": true, "predBGs": {"IOB": [110, 109]}}}`)}, now)
	require.JSONEq(t, `{"timestamp": "2026-11-23T05:57:00.000Z", "eventualBG": 100, "predBGs": {"IOB": [100, 99]},
		"moment": "2026-11-23T05:57:00.000Z"}`, asJSON(openAPS["lastSuggested"]))
	require.JSONEq(t, `{"timestamp": "2026-11-23T05:58:00.000Z", "eventualBG": 110, "received": true,
		"predBGs": {"IOB": [110, 109], "moment": "2026-11-23T05:58:00.000Z"}, "moment": "2026-11-23T05:58:00.000Z"}`, asJSON(openAPS["lastEnacted"]))
	require.JSONEq(t, `{"iob": 1.2, "time": "2026-11-23T05:58:00.000Z", "moment": "2026-11-23T05:58:00.000Z"}`, asJSON(openAPS["lastIOB"]))
	require.Equal(t, LoopState{"⌁", "enacted", "Enacted"}, openAPS["status"])

	loop := GetLoop(&sitedb.DeviceReport{LoopName: new("loop"), LoopStatusJson: new(`{"name": "Loop", "timestamp": "2026-11-23T05:59:00Z",
		"predicted": {"values": [100, 101]}, "enacted": {"timestamp": "2026-11-23T05:58:30Z", "rate": 0.4, "received": true}}`)}, now)
	require.JSONEq(t, `{"name": "Loop", "timestamp": "2026-11-23T05:59:00Z", "predicted": {"values": [100, 101]},
		"enacted": {"timestamp": "2026-11-23T05:58:30Z", "rate": 0.4, "received": true}, "moment": "2026-11-23T05:59:00.000Z"}`, asJSON(loop["lastLoop"]))
	require.JSONEq(t, `{"timestamp": "2026-11-23T05:58:30Z", "rate": 0.4, "received": true, "moment": "2026-11-23T05:58:30.000Z"}`, asJSON(loop["lastEnacted"]))
	require.Equal(t, LoopState{"⌁", "enacted", "Enacted"}, loop["display"])
	require.Nil(t, GetOpenAPS(&sitedb.DeviceReport{LoopName: new("loop"), LoopStatusJson: new(`{}`)}, now)["lastSuggested"])
}
