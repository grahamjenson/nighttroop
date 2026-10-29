package glucose

import (
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/sql/sitedb"
)

// reading is a reading of mgdl at seconds, written by identity, of type sgv unless it is mbg
func reading(id string, seconds int64, mgdl float64, identity string, readingType ...string) *sitedb.GlucoseReading {
	row := &sitedb.GlucoseReading{ID: id, OccurredAt: seconds * 1000, Mgdl: mgdl, CreatedByIdentityName: identity, ReadingType: "sgv"}
	if len(readingType) > 0 {
		row.ReadingType = readingType[0]
	}
	return row
}

func TestCleanReadings(t *testing.T) {
	type kept struct {
		id     string
		source string
		others []Other
	}
	tests := []struct {
		name string
		rows []*sitedb.GlucoseReading
		// want are the readings kept, newest first
		want []kept
	}{
		{"none", nil, []kept{}},
		{
			"the same glucose from two syncs is kept from Nightscout over Dexcom",
			[]*sitedb.GlucoseReading{reading("d", 10, 120, "dexcom_sync"), reading("n", 40, 122, "nightscout_sync")},
			[]kept{{"n", Nightscout, []Other{{"d", 120, Dexcom}}}},
		},
		{
			"a reading sent directly is kept over every sync's, and Libre's is kept last of them",
			[]*sitedb.GlucoseReading{reading("l", 0, 100, "libre_sync"), reading("p", 5, 101, "pwd_phone"), reading("n", 8, 99, "nightscout_sync")},
			[]kept{{"p", Direct, []Other{{"n", 99, Nightscout}, {"l", 100, Libre}}}},
		},
		{
			"a meter's reading is kept over all of them",
			[]*sitedb.GlucoseReading{reading("p", 0, 100, "pwd_phone"), reading("m", 30, 105, "admin", "mbg")},
			[]kept{{"m", Direct, []Other{{"p", 100, Direct}}}},
		},
		{
			"readings more than a minute apart are both kept",
			[]*sitedb.GlucoseReading{reading("a", 0, 100, "dexcom_sync"), reading("b", 61, 100, "nightscout_sync")},
			[]kept{{"b", Nightscout, nil}, {"a", Dexcom, nil}},
		},
		{
			"readings more than 10 mg/dL apart, as from two CGMs, are both kept",
			[]*sitedb.GlucoseReading{reading("a", 0, 100, "dexcom_sync"), reading("b", 10, 111, "libre_sync")},
			[]kept{{"b", Libre, nil}, {"a", Dexcom, nil}},
		},
		{
			"a reading is the same as the first of its glucose, not the last, so a slow rise is not one reading",
			[]*sitedb.GlucoseReading{reading("a", 0, 100, "p"), reading("b", 30, 108, "p"), reading("c", 50, 116, "p")},
			[]kept{{"c", Direct, nil}, {"a", Direct, []Other{{"b", 108, Direct}}}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := []kept{}
			for _, c := range CleanReadings(tt.rows) {
				got = append(got, kept{c.Row.ID, c.Source, c.Others})
			}
			require.Equal(t, tt.want, got)
		})
	}
}

func TestCleanReadingsTrend(t *testing.T) {
	// Searches give the newest first, and the duplicate is left out of the trend
	cleaned := CleanReadings([]*sitedb.GlucoseReading{
		reading("c", 600, 130, "p"), reading("b-copy", 310, 121, "dexcom_sync"), reading("b", 300, 120, "p"), reading("a", 0, 100, "p"),
	})
	require.Len(t, cleaned, 3)
	require.Nil(t, cleaned[2].DiffPer5mMgdl, "the oldest has nothing to change from")
	require.Equal(t, 20.0, *cleaned[1].DiffPer5mMgdl)
	require.Equal(t, 10.0, *cleaned[0].DiffPer5mMgdl)
}

func TestCleanReadingsTrendIsPerFiveMinutes(t *testing.T) {
	tests := []struct {
		name string
		rows []*sitedb.GlucoseReading
		want *float64
	}{
		{"a rise over 10 minutes is half of it every 5", []*sitedb.GlucoseReading{reading("a", 0, 100, "p"), reading("b", 600, 120, "p")}, new(10.0)},
		{"a fall over 2 minutes is two and a half times it every 5", []*sitedb.GlucoseReading{reading("a", 0, 100, "p"), reading("b", 120, 96, "p")}, new(-10.0)},
		{"two CGMs at the same time have no change between them", []*sitedb.GlucoseReading{reading("a", 0, 100, "p"), reading("b", 0, 140, "p")}, nil},
	}
	for _, tt := range tests {
		require.Equal(t, tt.want, CleanReadings(tt.rows)[0].DiffPer5mMgdl, tt.name)
	}
}
