package site_data_templ

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/models/sitemodels"
)

// Notes are kept as an app sent them, HTML too, so the site shows them escaped, as text
func TestEventGroupNotesAreEscaped(t *testing.T) {
	notes := `xDrip note <b>bold</b> & "quoted" <script>alert(1)</script>`
	var buf bytes.Buffer
	require.NoError(t, eventGroupRow(&sitemodels.SiteEventGroup{ID: "1", GroupType: "Note", Notes: &notes}).Render(context.Background(), &buf))
	require.NotContains(t, buf.String(), "<script>")
	require.NotContains(t, buf.String(), "<b>bold")
	require.Contains(t, buf.String(), "&lt;script&gt;alert(1)&lt;/script&gt;")
}
