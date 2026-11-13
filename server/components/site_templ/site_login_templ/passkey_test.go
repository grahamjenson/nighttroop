package site_login_templ

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

// The passkey script's own words, for a failure the browser gives no reason for, are the page's, from the locales,
// not written in the script
func TestPasskeyScriptFailsInThePagesWords(t *testing.T) {
	var html bytes.Buffer
	require.NoError(t, PasskeyScript().Render(context.Background(), &html))
	require.Contains(t, html.String(), `<script data-passkey-failed="The passkey did not work.">`)
	require.Contains(t, html.String(), "fail(e.message || failed)")
	require.Equal(t, 1, bytes.Count(html.Bytes(), []byte("The passkey did not work.")))
}
