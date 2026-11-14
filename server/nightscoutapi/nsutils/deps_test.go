package nsutils

import (
	"cmp"
	"fmt"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

// A delete's every find[$or] alternative must be one a delete takes, with a date or an id, before any is deleted
func TestParseDeleteQueries(t *testing.T) {
	tooMany := url.Values{}
	for i := range 11 {
		tooMany.Set(fmt.Sprintf("find[$or][%d][id][$eq]", i), fmt.Sprint("id", i))
	}
	tests := []struct {
		name       string
		values     url.Values
		queries    int
		ok         bool
		collection Collection
	}{
		{"trio carbs", url.Values{"find[id][$eq]": {"A1B2C3D4"}}, 1, true, ""},
		{"trio override", url.Values{"find[created_at][$eq]": {"2026-11-02T12:13:00.000Z"}, "find[eventType][$eq]": {"Exercise"}}, 1, true, ""},
		{"trio glucose", url.Values{"find[$or][0][id][$eq]": {"a"}, "find[$or][1][dateString][$eq]": {"2026-11-02T12:00:00.000Z"}}, 2, true, ""},
		{"no date or id", url.Values{"find[eventType][$eq]": {"Exercise"}}, 0, false, ""},
		{"unknown filter", url.Values{"find[id][$eq]": {"a"}, "find[colour][$eq]": {"blue"}}, 0, false, ""},
		{"unknown in an alternative", url.Values{"find[$or][0][id][$eq]": {"a"}, "find[$or][1][colour][$eq]": {"blue"}}, 0, false, ""},
		{"too many alternatives", tooMany, 0, false, ""},
		{"iaps carbs", url.Values{"find[carbs][$exists]": {"true"}, "find[creation_date][$eq]": {"2026-11-02T12:01:00.000Z"}}, 1, true, ""},
		{"iaps insulin", url.Values{"find[bolus][$exists]": {"true"}, "find[created_at][$eq]": {"2026-11-02T12:02:00.000Z"}}, 1, true, ""},
		{"iaps glucose", url.Values{"find[glucose][$exists]": {"true"}, "find[created_at][$eq]": {"2026-11-02T12:20:00.000Z"}}, 1, true, ""},
		{"iaps override", url.Values{"find[eventType]": {"Exercise"}, "find[enteredBy]": {"iAPS"}, "find[created_at][$gte]": {"2026-11-02T12:08:00.000Z"}, "find[created_at][$lte]": {"2026-11-02T12:18:00.000Z"}}, 1, true, ""},
		{"iaps newest override", url.Values{"find[eventType]": {"Exercise"}, "count": {"1"}, "find[enteredBy]": {"iAPS"}}, 1, true, ""},
		{"iaps announcements", url.Values{"find[eventType]": {"Announcement"}, "find[created_at][$gte]": {"2026-11-02T12:00:00.000Z"}}, 1, true, ""},
		{"iaps all overrides", url.Values{"find[eventType]": {"Exercise"}}, 0, false, ""},
		{"by who without a type", url.Values{"find[enteredBy]": {"iAPS"}, "count": {"1"}}, 0, false, ""},
		{"range of more than 2 days", url.Values{"find[created_at][$gte]": {"2026-11-01T12:00:00.000Z"}, "find[created_at][$lte]": {"2026-11-03T12:00:01.000Z"}}, 0, false, ""},
		{"range of 2 days", url.Values{"find[created_at][$gte]": {"2026-11-01T12:00:00.000Z"}, "find[created_at][$lte]": {"2026-11-03T12:00:00.000Z"}}, 1, true, ""},
		{"count of 0", url.Values{"find[eventType]": {"Exercise"}, "count": {"0"}}, 0, false, ""},
		{"count of 11", url.Values{"find[eventType]": {"Exercise"}, "count": {"11"}}, 0, false, ""},
		{"readings by a treatment's filter", url.Values{"find[bolus][$exists]": {"true"}, "find[created_at][$eq]": {"2026-11-02T12:02:00.000Z"}}, 0, false, Entries},
		{"readings by carbs", url.Values{"find[carbs][$exists]": {"true"}, "find[date][$gte]": {"1793000000000"}}, 0, false, Entries},
		{"readings by date", url.Values{"find[type]": {"sgv"}, "find[date][$gte]": {"1793000000000"}}, 1, true, Entries},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			queries, err := ParseDeleteQueries(tt.values, cmp.Or(tt.collection, Treatments))
			if !tt.ok {
				assert.NotNil(t, err)
				return
			}
			assert.Nil(t, err)
			assert.Len(t, queries, tt.queries)
		})
	}
}
