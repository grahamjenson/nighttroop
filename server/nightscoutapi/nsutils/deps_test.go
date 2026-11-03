package nsutils

import (
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
		name    string
		values  url.Values
		queries int
		ok      bool
	}{
		{"trio carbs", url.Values{"find[id][$eq]": {"A1B2C3D4"}}, 1, true},
		{"trio override", url.Values{"find[created_at][$eq]": {"2026-11-02T12:13:00.000Z"}, "find[eventType][$eq]": {"Exercise"}}, 1, true},
		{"trio glucose", url.Values{"find[$or][0][id][$eq]": {"a"}, "find[$or][1][dateString][$eq]": {"2026-11-02T12:00:00.000Z"}}, 2, true},
		{"no date or id", url.Values{"find[eventType][$eq]": {"Exercise"}}, 0, false},
		{"unknown filter", url.Values{"find[id][$eq]": {"a"}, "find[colour][$eq]": {"blue"}}, 0, false},
		{"unknown in an alternative", url.Values{"find[$or][0][id][$eq]": {"a"}, "find[$or][1][colour][$eq]": {"blue"}}, 0, false},
		{"too many alternatives", tooMany, 0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			queries, err := ParseDeleteQueries(tt.values)
			if !tt.ok {
				assert.NotNil(t, err)
				return
			}
			assert.Nil(t, err)
			assert.Len(t, queries, tt.queries)
		})
	}
}
