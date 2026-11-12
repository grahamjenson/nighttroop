package types_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/utils/types"
)

func TestDateStringUnmarshalNumbers(t *testing.T) {
	// A number has no time zone, so it must not take the machine's; run as if in New Zealand to show it
	local := time.Local
	time.Local = time.FixedZone("NZST", 12*60*60)
	t.Cleanup(func() { time.Local = local })

	tests := []struct {
		json string
		want string
	}{
		{json: `1719300000000`, want: "2024-06-25T07:20:00Z"},
		{json: `1719300000000.5`, want: "2024-06-25T07:20:00Z"},
		{json: `1.7193e12`, want: "2024-06-25T07:20:00Z"},
		{json: `"2024-06-25T07:20:00Z"`, want: "2024-06-25T07:20:00Z"},
	}
	for _, test := range tests {
		t.Run(test.json, func(t *testing.T) {
			var date types.DateString
			require.NoError(t, json.Unmarshal([]byte(test.json), &date))
			require.Equal(t, test.want, date.String())
		})
	}
}

// A time to the minute, as some clients find treatments since, is UTC without a zone, or in its zone with one
func TestStringToTimeToTheMinute(t *testing.T) {
	want := time.Date(2025, 3, 4, 10, 30, 0, 0, time.UTC)
	for _, value := range []string{"2025-03-04T10:30", "2025-03-04T23:30+13:00", "2025-03-04T10:30Z"} {
		got, err := types.StringToTime(value)
		require.NoError(t, err, value)
		require.True(t, want.Equal(got), "%s is %s", value, got)
	}
}

// A created_at is written as JavaScript's toISOString writes it, as Nightscout does: in UTC with three fractional
// digits, whatever it was sent with
func TestISOString(t *testing.T) {
	tests := []struct{ at, want string }{
		{"2026-11-12T08:01:23.940Z", "2026-11-12T08:01:23.940Z"},
		{"2026-11-12T08:01:23Z", "2026-11-12T08:01:23.000Z"},
		{"2026-11-12T08:01:23.123456Z", "2026-11-12T08:01:23.123Z"},
		{"2026-11-12T21:01:23.5+13:00", "2026-11-12T08:01:23.500Z"},
	}
	for _, test := range tests {
		t.Run(test.at, func(t *testing.T) {
			at, err := time.Parse(time.RFC3339Nano, test.at)
			require.NoError(t, err)
			require.Equal(t, test.want, types.ISOString(at))
		})
	}
}

// A dateString is written with three fractional digits, as toISOString does, in its UTC offset, which it gives
func TestISODateStringFromInt64AndUTCOffset(t *testing.T) {
	zero, nz, newfoundland := int64(0), int64(780), int64(-210)
	tests := []struct {
		at     int64
		offset *int64
		want   string
	}{
		{1794507802510, nil, "2026-11-12T18:23:22.510Z"},
		{1794507802000, nil, "2026-11-12T18:23:22.000Z"},
		{1794507802940, &zero, "2026-11-12T18:23:22.940Z"},
		{1794507802500, &nz, "2026-11-13T07:23:22.500+13:00"},
		{1794507802000, &newfoundland, "2026-11-12T14:53:22.000-03:30"},
	}
	for _, test := range tests {
		t.Run(test.want, func(t *testing.T) {
			got := types.ISODateStringFromInt64AndUTCOffset(test.at, test.offset)
			require.Equal(t, test.want, string(got))
			require.Equal(t, test.at, got.UnixMilli())
		})
	}
}
