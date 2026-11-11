package heading

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/utils/httputils"
)

// TestPageNamesTheTab checks a page's heading also titles the browser's tab, for the site or the control site
func TestPageNamesTheTab(t *testing.T) {
	tests := []struct {
		name  string
		ctx   context.Context
		title string
	}{
		{"site", context.Background(), "<title>Glucose Readings · Nighttroop</title>"},
		{"control", httputils.CreateControlContext(context.Background(), nil, "127.0.0.1", httputils.HTTPActionRead), "<title>Glucose Readings · Nighttroop Control</title>"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			require.NoError(t, Page("Glucose Readings").Render(tt.ctx, &buf))
			require.Contains(t, buf.String(), tt.title)
			require.Contains(t, buf.String(), `<h1 class="text-2xl font-bold">Glucose Readings</h1>`)
		})
	}
}
