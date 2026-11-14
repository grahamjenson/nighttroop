package types_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/utils/types"
)

// A duration is read from duration in minutes, possibly with a fraction, or durationInMilliseconds, or both: minutes
// sent are kept whole, and to the millisecond when they have a fraction and no milliseconds are sent; milliseconds sent
// are kept as sent, and in the nearest minutes when no minutes are sent
func TestSentDuration(t *testing.T) {
	tests := []struct {
		name    string
		sent    string
		minutes int64
		ms      *int64
		// fraction is whether the minutes were sent with a fraction
		fraction bool
	}{
		{"whole minutes", `{"duration": 30}`, 30, nil, false},
		{"minutes as a string", `{"duration": "30"}`, 30, nil, false},
		{"Loop's minutes with a fraction", `{"duration": 1.6666666666666667}`, 1, new(int64(100000)), true},
		{"milliseconds", `{"durationInMilliseconds": 1799000}`, 30, new(int64(1799000)), false},
		{"AAPS's minutes and milliseconds", `{"duration": 5, "durationInMilliseconds": 359587}`, 5, new(int64(359587)), false},
		{"milliseconds over a fraction", `{"duration": 5.5, "durationInMilliseconds": 359587}`, 5, new(int64(359587)), true},
		{"none of 0 minutes", `{"duration": 0, "durationInMilliseconds": 43491}`, 0, new(int64(43491)), false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var sent types.SentDuration
			require.NoError(t, json.Unmarshal([]byte(test.sent), &sent))
			d := sent.Duration()
			require.Equal(t, test.minutes, d.Minutes)
			require.Equal(t, test.ms, d.Ms)
			require.Equal(t, test.fraction, d.Fraction)
		})
	}
	var none types.SentDuration
	require.NoError(t, json.Unmarshal([]byte(`{"duration": null}`), &none))
	require.Nil(t, none.Duration())
}

// A duration is written in both minutes and milliseconds. Its milliseconds are those kept while they are within a
// minute of its minutes, which the site may change. Its minutes have their fraction on API v1 when they were sent with
// one, as Loop reads them, and are whole on API v3, as AAPS reads them, and as Nightscout gives back the minutes sent.
func TestDurationFields(t *testing.T) {
	tests := []struct {
		name   string
		d      *types.Duration
		v1, v3 string
	}{
		{"none", nil, `{}`, `{}`},
		{"whole minutes", types.DurationOfMinutes(new(int64(30))), `{"duration":30,"durationInMilliseconds":1800000}`, `{"duration":30,"durationInMilliseconds":1800000}`},
		{"AAPS's whole minutes and milliseconds", &types.Duration{Minutes: 5, Ms: new(int64(359587))}, `{"duration":5,"durationInMilliseconds":359587}`, `{"duration":5,"durationInMilliseconds":359587}`},
		{"a fraction of a minute sent", &types.Duration{Minutes: 2, Ms: new(int64(150000)), Fraction: true}, `{"duration":2.5,"durationInMilliseconds":150000}`, `{"duration":2,"durationInMilliseconds":150000}`},
		{"Loop's bolus", &types.Duration{Minutes: 1, Ms: new(int64(100000)), Fraction: true}, `{"duration":1.6666666666666667,"durationInMilliseconds":100000}`, `{"duration":1,"durationInMilliseconds":100000}`},
		{"rounded up to its minutes", types.DurationOfMs(new(int64(1799000))), `{"duration":30,"durationInMilliseconds":1799000}`, `{"duration":30,"durationInMilliseconds":1799000}`},
		{"minutes changed on the site", (&types.Duration{Minutes: 90}).WithMs(new(int64(359587)), true), `{"duration":90,"durationInMilliseconds":5400000}`, `{"duration":90,"durationInMilliseconds":5400000}`},
		{"AAPS's under a minute", &types.Duration{Minutes: 0, Ms: new(int64(43491))}, `{"duration":0,"durationInMilliseconds":43491}`, `{"duration":0,"durationInMilliseconds":43491}`},
		{"under a minute sent", &types.Duration{Minutes: 0, Ms: new(int64(43491)), Fraction: true}, `{"duration":0.72485,"durationInMilliseconds":43491}`, `{"duration":0,"durationInMilliseconds":43491}`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			v1, err := json.Marshal(test.d.Fields())
			require.NoError(t, err)
			require.Equal(t, test.v1, string(v1))
			v3, err := json.Marshal(test.d.InWholeMinutes().Fields())
			require.NoError(t, err)
			require.Equal(t, test.v3, string(v3))
		})
	}
}

// A duration kept in an event's minutes takes the milliseconds kept with it and whether its minutes were sent with a
// fraction, and one of no minutes is of those milliseconds, in the nearest minutes
func TestDurationWithMs(t *testing.T) {
	require.Nil(t, (*types.Duration)(nil).WithMs(nil, false))
	require.Equal(t, &types.Duration{Minutes: 6, Ms: new(int64(359587))}, (*types.Duration)(nil).WithMs(new(int64(359587)), false))
	require.Equal(t, &types.Duration{Minutes: 5}, types.DurationOfMinutes(new(int64(5))).WithMs(nil, true))
	require.Equal(t, &types.Duration{Minutes: 5, Ms: new(int64(359587))}, types.DurationOfMinutes(new(int64(5))).WithMs(new(int64(359587)), false))
	require.Equal(t, &types.Duration{Minutes: 2, Ms: new(int64(150000)), Fraction: true}, types.DurationOfMinutes(new(int64(2))).WithMs(new(int64(150000)), true))
}
