package nscollections

import (
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/data/sql/sitedb"
	"nighttroop.t1d.nz/models/nightscoutmodels"
	"nighttroop.t1d.nz/models/sitemodels"
	"nighttroop.t1d.nz/utils/httputils"
)

// Every type of treatment keeps the kinds of event the site lets its group have, each a kind the API can write
func TestTreatmentTypeEvents(t *testing.T) {
	for _, eventType := range nightscoutmodels.TreatmentEventTypes {
		kinds, ok := sitemodels.EventGroupTypeKinds[eventType]
		require.True(t, ok, eventType)
		require.Len(t, treatmentTypeEvents[eventType], len(kinds), eventType)
		for _, event := range treatmentTypeEvents[eventType] {
			require.NotNil(t, event, eventType)
		}
	}
}

// A treatment's duration sent in milliseconds is given back while it is within a minute of the minutes its event keeps,
// and from those minutes once they are changed on the site
func TestDurationInMilliseconds(t *testing.T) {
	tests := []struct {
		name     string
		ms       *int64
		minutes  int64
		expected int64
	}{
		{"ended early, to the millisecond", new(int64(359587)), 5, 359587},
		{"rounded up to the minutes", new(int64(1799000)), 30, 1799000},
		{"none sent, from the minutes", nil, 60, 3600000},
		{"minutes changed on the site", new(int64(359587)), 90, 5400000},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			bundle := sitedb.EventGroupsBundle{
				EventGroup: sitedb.EventGroup{GroupType: string(nightscoutmodels.TemporaryTarget), DurationMs: test.ms},
				TempTarget: &sitedb.TempTarget{DurationMinutes: test.minutes},
			}
			site := &httputils.SiteContext{SiteID: "abcdefgh", SiteModel: &controldb.Site{ID: "abcdefgh"}}
			treatment := DBModelToTreatment(&bundle, site)
			require.Equal(t, test.expected, treatment.DurationInMilliseconds.Int64())
			require.Equal(t, test.minutes, treatment.DurationMinutes.Int64())
		})
	}
}
