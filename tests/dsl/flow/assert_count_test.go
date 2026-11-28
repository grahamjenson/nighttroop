package flow

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/require"
)

// The list says it is empty, and shows its 3 rows 300ms after it loads, as an HTMX swap does after a click clears a
// filter that matched nothing
const swappingPage = `<html><body><table><tbody id="rows"><tr id="none"><td>No day matches the filter.</td></tr></tbody></table><script>
setTimeout(() => { document.getElementById("rows").innerHTML = '<tr class="day"><td>1</td></tr><tr class="day"><td>2</td></tr><tr class="day"><td>3</td></tr>' }, 300)
</script></body></html>`

// A count that is not yet the expected one is counted again until the page shows it, and one that is expected passes
// at once
func TestCountAssertWaitsForThePage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(swappingPage))
	}))
	t.Cleanup(server.Close)
	pw, err := playwright.Run()
	require.NoError(t, err)
	t.Cleanup(func() { _ = pw.Stop() })
	browser, err := pw.Chromium.Launch()
	require.NoError(t, err)
	t.Cleanup(func() { _ = browser.Close() })

	page, err := browser.NewPage()
	require.NoError(t, err)
	_, err = page.Goto(server.URL)
	require.NoError(t, err)
	require.NoError(t, (&AssertCount{Locator: "#none", Count: 1}).Execute(t, map[string]any{}, page, nil))
	require.NoError(t, (&AssertCount{Locator: "#rows tr.day", Count: 3}).Execute(t, map[string]any{}, page, nil))

	started := time.Now()
	require.NoError(t, (&AssertCount{Locator: "#rows tr.day", Count: 3}).Execute(t, map[string]any{}, page, nil))
	require.NoError(t, (&AssertCount{Locator: "#none", Count: 0}).Execute(t, map[string]any{}, page, nil))
	require.Less(t, time.Since(started), time.Second, "the expected count is not waited for")
}
