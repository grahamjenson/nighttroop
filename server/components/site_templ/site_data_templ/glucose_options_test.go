package site_data_templ

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/data/sql/controldb"
	"nighttroop.t1d.nz/utils/httputils"
)

// TestGlucoseOptionsReadAsTheTable checks adding a reading offers its type and trend as the table shows them,
// while still sending Nightscout's values
func TestGlucoseOptionsReadAsTheTable(t *testing.T) {
	ctx := httputils.CreateSiteContext(context.Background(), nil, nil, nil, &controldb.Site{}, "site", "", 0, 0, "")
	ctx = httputils.SetIdentityModelInSiteContext(ctx, &controldb.SiteIdentity{})
	var buf bytes.Buffer
	require.NoError(t, GlucoseReadingNew().Render(ctx, &buf))
	html := buf.String()
	for _, option := range []string{
		`<option value="mbg">Meter</option>`,
		`<option value="sgv">Sensor</option>`,
		`<option value="FortyFiveUp">↗ FortyFiveUp</option>`,
	} {
		require.Contains(t, html, option)
	}
}
