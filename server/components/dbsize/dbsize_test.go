package dbsize

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/models/controlmodels"
)

// The DB size bar is as wide as the share of its limit the site uses, at most all of it, in the theme's colour, amber
// nearing the limit and red over it; a size not measured has no bar
func TestDBSizeBar(t *testing.T) {
	size := func(kb int64) *int64 { return &kb }
	tables := []*controlmodels.SiteTable{{Name: "glucose_readings", Bytes: 3 << 20}, {Name: "treatments", Bytes: 1 << 20}}
	for _, test := range []struct {
		name   string
		sizeKB *int64
		tables []*controlmodels.SiteTable
		want   []string
		bar    bool
	}{
		{"half", size(1500), tables, []string{`style="width: 50%;"`, "50% of its limit used", `id="control--site--db-table--glucose_readings"`, "3.00 MB · 75%"}, true},
		{"nearing it", size(2700), nil, []string{`style="width: 90%;"`, "text-amber-800", "90% of its limit used"}, true},
		{"over it", size(4500), nil, []string{`style="width: 100%;"`, "text-red-600", "150% of its limit used"}, true},
		{"measured by its tables", nil, tables, []string{">4.00</span>", `style="width: 100%;"`, "136% of its limit used", `id="control--site--db-part--treatments"`}, true},
		{"not measured", nil, nil, []string{"Not measured yet", "The resource check measures each site&#39;s DB"}, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			var buf bytes.Buffer
			require.NoError(t, Panel("control--site", test.sizeKB, size(3000), test.tables).Render(context.Background(), &buf))
			for _, want := range test.want {
				require.Contains(t, buf.String(), want)
			}
			require.Equal(t, test.bar, bytes.Contains(buf.Bytes(), []byte(`id="control--site--db-used"`)))
			require.Equal(t, test.tables != nil, bytes.Contains(buf.Bytes(), []byte(`id="control--site--db-tables"`)))
		})
	}
}

// The largest tables are shaded darkest to lightest, by their share of the DB, and the rest are Other
func TestDBParts(t *testing.T) {
	tables := []*controlmodels.SiteTable{}
	for i, bytes := range []int64{40, 20, 15, 10, 5, 6, 4} {
		tables = append(tables, &controlmodels.SiteTable{Name: fmt.Sprintf("t%d", i), Bytes: bytes, Records: bytes * 100})
	}
	parts := dbParts(tables)
	require.Len(t, parts, 6)
	require.Equal(t, dbPart{Name: "t0", Bytes: 40, Records: 4000, Percent: 40, Class: "bg-brand-night"}, parts[0])
	require.Equal(t, dbPart{Name: "t4", Bytes: 5, Records: 500, Percent: 5, Class: "bg-brand-night/25"}, parts[4])
	// Other's records are its tables' together
	require.Equal(t, dbPart{Name: "Other", Bytes: 10, Records: 1000, Percent: 10, Class: "bg-gray-300"}, parts[5])
	require.Equal(t, "4.0k · 0.00 MB · 40%", keyText(parts[0]))
	require.Len(t, dbParts(tables[:2]), 2)
	require.Empty(t, dbParts(nil))
	require.Empty(t, dbParts([]*controlmodels.SiteTable{{Name: "empty"}}))
}
