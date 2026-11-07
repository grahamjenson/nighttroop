package site_summary_templ

import (
	"bytes"
	"context"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/models/sitemodels"
)

// TestEmptyGlucoseLinksToSyncSettings checks a summary with no readings says how to get some, linking to Sync
func TestEmptyGlucoseLinksToSyncSettings(t *testing.T) {
	var buf bytes.Buffer
	require.NoError(t, glucoseHero(&sitemodels.Summary{}).Render(context.Background(), &buf))
	html := buf.String()
	require.Contains(t, html, `id="site--summary--sync-link" href="/settings/sync" hx-get="/settings/sync" hx-target="#main-content"`)
	text := strings.Join(strings.Fields(regexp.MustCompile(`<[^>]*>`).ReplaceAllString(html, " ")), " ")
	require.Contains(t, text, "Upload from your phone, or connect a source in Sync .")
}
