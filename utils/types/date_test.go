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
