package site_data_templ

import (
	"bytes"
	"context"
	"regexp"
	"strconv"
	"testing"

	"github.com/a-h/templ"
	"github.com/stretchr/testify/require"
)

// TestEmptyRowsSpanTheTable checks a table with no rows says so across all its columns
func TestEmptyRowsSpanTheTable(t *testing.T) {
	tests := []struct {
		name  string
		table templ.Component
	}{
		{"activities", ActivitiesTable(nil, 0, 25)},
		{"component events", ComponentEventsTable(nil, 0, 25)},
		{"event groups", EventGroupsTable(nil, 0, 25)},
		{"food intakes", FoodIntakesTable(nil, 0, 25)},
		{"insulin doses", InsulinDosesTable(nil, 0, 25)},
		{"notification events", NotificationEventsTable(nil, 0, 25)},
		{"profile switches", ProfileSwitchesTable(nil, 0, 25)},
		{"temp basals", TempBasalsTable(nil, 0, 25)},
	}
	colspan := regexp.MustCompile(`<td colspan="(\d+)"`)
	head := regexp.MustCompile(`<th[\s>]`)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			require.NoError(t, tt.table.Render(context.Background(), &buf))
			html := buf.String()
			match := colspan.FindStringSubmatch(html)
			require.NotNil(t, match, html)
			span, _ := strconv.Atoi(match[1])
			require.Equal(t, len(head.FindAllString(html, -1)), span, "the empty row should span every column")
		})
	}
}
