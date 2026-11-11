package listpage

import (
	"bytes"
	"context"
	"regexp"
	"strconv"
	"testing"

	"github.com/a-h/templ"
	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/server/components/ui/table"
)

// TestEmptyRowSpansTheTable checks a table's row saying it has no rows spans every column, is there for rows added and
// deleted in place, and for Static rows only when there are none
func TestEmptyRowSpansTheTable(t *testing.T) {
	columns := []table.Column{{Label: "Time"}, {Label: "Units", Numeric: true}, {Actions: true}}
	row := templ.Raw(`<tr id="row"><td>1</td><td>2</td><td></td></tr>`)
	tests := []struct {
		name      string
		props     TableProps
		rows      bool
		wantEmpty bool
	}{
		{"always there, with no rows", TableProps{EmptyID: "x--empty"}, false, true},
		{"always there, with rows", TableProps{EmptyID: "x--empty"}, true, true},
		{"static, with no rows", TableProps{EmptyID: "x--empty", Static: true}, false, true},
		{"static, with rows", TableProps{EmptyID: "x--empty", Static: true, Rows: 1}, true, false},
	}
	colspan := regexp.MustCompile(`<td colspan="(\d+)"`)
	head := regexp.MustCompile(`<th[\s>]`)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.props.Columns, tt.props.Empty = columns, "Nothing here."
			var buf bytes.Buffer
			ctx := context.Background()
			if tt.rows {
				ctx = templ.WithChildren(ctx, row)
			}
			require.NoError(t, Table(tt.props).Render(ctx, &buf))
			html := buf.String()
			require.Len(t, head.FindAllString(html, -1), len(columns))
			require.Contains(t, html, `<span class="sr-only">Actions</span>`)
			match := colspan.FindStringSubmatch(html)
			if !tt.wantEmpty {
				require.Nil(t, match, html)
				require.NotContains(t, html, "x--empty")
				return
			}
			require.NotNil(t, match, html)
			span, _ := strconv.Atoi(match[1])
			require.Equal(t, len(columns), span, "the empty row should span every column")
			require.Contains(t, html, `id="x--empty"`)
		})
	}
}
