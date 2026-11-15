package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/utils/types"
)

// aliases are a collection's time aliases, as a treatment's: date first, then eventTime and created_at, then a
// timestamp written on API v1 only, and a sysTime only read
var aliases = types.TimeAliases{{Name: "date", Millis: true}, {Name: "eventTime", Write: types.WriteNever},
	{Name: "created_at"}, {Name: "dateString"}, {Name: "mills", Millis: true}, {Name: "timestamp", Write: types.WriteV1},
	{Name: "sysTime", Write: types.WriteNever}}

// Each alias is read on its own, in milliseconds or seconds for a number and in its zone for a string, which gives the
// offset unless it is UTC; the first sent in the aliases' order is the time, and one sent after it in milliseconds in
// the same second gives it its milliseconds
func TestTimeAliasesRead(t *testing.T) {
	nz := int64(780)
	tests := []struct {
		name   string
		sent   string
		date   int64
		offset *int64
	}{
		{"date", `{"date": 1794512149419}`, 1794512149419, nil},
		{"date in seconds", `{"date": 1794512149}`, 1794512149000, nil},
		{"date as a string", `{"date": "1794512149419.0"}`, 1794512149419, nil},
		{"eventTime", `{"eventTime": "2026-11-12T19:35:49.419Z"}`, 1794512149419, nil},
		{"created_at", `{"created_at": "2026-11-12T19:35:49.419Z"}`, 1794512149419, nil},
		{"created_at with its zone", `{"created_at": "2026-11-13T08:35:49.419+13:00"}`, 1794512149419, &nz},
		{"dateString", `{"dateString": "2026-11-13T08:35:49.419+13:00"}`, 1794512149419, &nz},
		{"mills", `{"mills": 1794512149419}`, 1794512149419, nil},
		{"timestamp as a string", `{"timestamp": "2026-11-12T19:35:49.419Z"}`, 1794512149419, nil},
		{"timestamp in ms", `{"timestamp": 1794512149419}`, 1794512149419, nil},
		{"sysTime", `{"sysTime": "2026-11-12T19:35:49.419Z"}`, 1794512149419, nil},
		{"date before created_at", `{"created_at": "2026-11-12T19:35:00Z", "date": 1794512149419}`, 1794512149419, nil},
		{"eventTime before created_at", `{"created_at": "2026-11-12T19:35:00Z", "eventTime": "2026-11-12T19:35:49.419Z"}`, 1794512149419, nil},
		{"the offset of the first string", `{"date": 1794512149419, "eventTime": "2026-11-12T19:35:49.419Z", "created_at": "2026-11-13T08:35:49.419+13:00"}`, 1794512149419, nil},
		{"the offset of a later string", `{"date": 1794512149419, "created_at": "2026-11-13T08:35:49.419+13:00"}`, 1794512149419, &nz},
		{"xDrip+'s timestamp gives the milliseconds", `{"created_at": "2026-11-12T19:35:49Z", "timestamp": 1794512149419}`, 1794512149419, nil},
		{"a timestamp in another second does not", `{"created_at": "2026-11-12T19:35:49Z", "timestamp": 1794512150419}`, 1794512149000, nil},
		{"a timestamp string does not", `{"created_at": "2026-11-12T19:35:49Z", "timestamp": "2026-11-12T19:35:49.419Z"}`, 1794512149000, nil},
		{"one not given back is read only when it is a time", `{"created_at": "2026-11-12T19:35:49Z", "timestamp": "soon"}`, 1794512149000, nil},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			date, offset, err := aliases.Read([]byte(test.sent))
			require.NoError(t, err)
			require.Equal(t, test.date, date.Int64())
			require.Equal(t, test.offset, offset.Int64p())
		})
	}
}

// The offset sent is kept, a document with no time has none, and one not a time refuses it
func TestTimeAliasesReadOffsetAndErrors(t *testing.T) {
	date, offset, err := aliases.Read([]byte(`{"created_at": "2026-11-13T08:35:49.419+13:00", "utcOffset": "-210"}`))
	require.NoError(t, err)
	require.Equal(t, int64(1794512149419), date.Int64())
	require.Equal(t, int64(-210), offset.Int64())

	date, offset, err = aliases.Read([]byte(`{"notes": "no time"}`))
	require.NoError(t, err)
	require.Nil(t, date)
	require.Nil(t, offset)

	_, _, err = aliases.Read([]byte(`{"date": 1794512149419, "dateString": "notARealTime"}`))
	require.ErrorContains(t, err, "date string")
}

// Each alias is written from when it happened, in milliseconds or as toISOString writes it, in UTC with three
// fractional digits, whatever its offset; one Nighttroop makes up is left out of API v3, and one only read out of both
func TestTimeAliasesWritten(t *testing.T) {
	v1 := [][2]string{{"date", "1794512149000"}, {"created_at", `"2026-11-12T19:35:49.000Z"`},
		{"dateString", `"2026-11-12T19:35:49.000Z"`}, {"mills", "1794512149000"}, {"timestamp", `"2026-11-12T19:35:49.000Z"`}}
	require.Equal(t, v1, aliases.Written(1794512149000, false))
	require.Equal(t, v1[:4], aliases.Written(1794512149000, true))
}
