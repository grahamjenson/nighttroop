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
	}{
		{"whole minutes", `{"duration": 30}`, 30, nil},
		{"minutes as a string", `{"duration": "30"}`, 30, nil},
		{"Loop's minutes with a fraction", `{"duration": 1.6666666666666667}`, 1, new(int64(100000))},
		{"milliseconds", `{"durationInMilliseconds": 1799000}`, 30, new(int64(1799000))},
		{"AAPS's minutes and milliseconds", `{"duration": 5, "durationInMilliseconds": 359587}`, 5, new(int64(359587))},
		{"milliseconds over a fraction", `{"duration": 5.5, "durationInMilliseconds": 359587}`, 5, new(int64(359587))},
		{"none of 0 minutes", `{"duration": 0, "durationInMilliseconds": 43491}`, 0, new(int64(43491))},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var sent types.SentDuration
			require.NoError(t, json.Unmarshal([]byte(test.sent), &sent))
			d := sent.Duration()
			require.Equal(t, test.minutes, d.Minutes)
			require.Equal(t, test.ms, d.Ms)
		})
	}
	var none types.SentDuration
	require.NoError(t, json.Unmarshal([]byte(`{"duration": null}`), &none))
	require.Nil(t, none.Duration())
}

// A duration is written in both minutes and milliseconds. Its milliseconds are those kept while they are within a
// minute of its minutes, which the site may change. Its minutes have their fraction on API v1, as Loop reads them, and
// are whole on API v3, as AAPS reads them.
func TestDurationFields(t *testing.T) {
	tests := []struct {
		name   string
		d      *types.Duration
		v1, v3 string
	}{
		{"none", nil, `{}`, `{}`},
		{"whole minutes", types.DurationOfMinutes(new(int64(30))), `{"duration":30,"durationInMilliseconds":1800000}`, `{"duration":30,"durationInMilliseconds":1800000}`},
		{"a fraction of a minute", &types.Duration{Minutes: 5, Ms: new(int64(359587))}, `{"duration":5.993116666666666,"durationInMilliseconds":359587}`, `{"duration":5,"durationInMilliseconds":359587}`},
		{"Loop's bolus", &types.Duration{Minutes: 1, Ms: new(int64(100000))}, `{"duration":1.6666666666666667,"durationInMilliseconds":100000}`, `{"duration":1,"durationInMilliseconds":100000}`},
		{"rounded up to its minutes", types.DurationOfMs(new(int64(1799000))), `{"duration":30,"durationInMilliseconds":1799000}`, `{"duration":30,"durationInMilliseconds":1799000}`},
		{"minutes changed on the site", (&types.Duration{Minutes: 90}).WithMs(new(int64(359587))), `{"duration":90,"durationInMilliseconds":5400000}`, `{"duration":90,"durationInMilliseconds":5400000}`},
		{"under a minute", &types.Duration{Minutes: 0, Ms: new(int64(43491))}, `{"duration":0.72485,"durationInMilliseconds":43491}`, `{"duration":0,"durationInMilliseconds":43491}`},
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

// A duration kept in an event's minutes takes the milliseconds kept with it, and one of no minutes is of those
// milliseconds, in the nearest minutes
func TestDurationWithMs(t *testing.T) {
	require.Nil(t, (*types.Duration)(nil).WithMs(nil))
	require.Equal(t, &types.Duration{Minutes: 6, Ms: new(int64(359587))}, (*types.Duration)(nil).WithMs(new(int64(359587))))
	require.Equal(t, &types.Duration{Minutes: 5}, types.DurationOfMinutes(new(int64(5))).WithMs(nil))
	require.Equal(t, &types.Duration{Minutes: 5, Ms: new(int64(359587))}, types.DurationOfMinutes(new(int64(5))).WithMs(new(int64(359587))))
}
