package head_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/server/assets/static"
	"nighttroop.t1d.nz/server/components/head"
)

// Every page's head has the style and htmx at their versioned URLs, and a signed in page's has its components' scripts
func TestHeadLinksStyleAndScripts(t *testing.T) {
	var common, components bytes.Buffer
	require.NoError(t, head.Common().Render(context.Background(), &common))
	require.NoError(t, head.Components().Render(context.Background(), &components))

	for _, path := range []string{"/style_gen.css", "/htmx.js", "/qrcode.js"} {
		require.Contains(t, common.String(), static.URL(path))
	}
	require.Contains(t, common.String(), `name="viewport"`)
	for _, script := range []string{"input", "label", "dropdown", "inputotp", "checkbox", "selectbox", "toast"} {
		require.Contains(t, components.String(), "/"+script+".min.js")
	}
}
