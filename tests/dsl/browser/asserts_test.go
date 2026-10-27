package browser

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/require"
)

// The page shows its heading half a second after it loads, as a page swapped in after a login does
const latePage = `<html><body><script>
setTimeout(() => { document.body.innerHTML = '<h1 id="late">Nighttroop</h1>' }, 500)
</script></body></html>`

func TestVisibleAssertsWaitForThePage(t *testing.T) {
	pw, err := playwright.Run()
	require.NoError(t, err)
	t.Cleanup(func() { _ = pw.Stop() })
	browser, err := pw.Chromium.Launch()
	require.NoError(t, err)
	t.Cleanup(func() { _ = browser.Close() })

	tests := []struct {
		name   string
		assert interface {
			Execute(*testing.T, map[string]any, playwright.Page, []playwright.Response) error
		}
	}{
		{"ASSERT_VISIBLE id", &AssertVisible{Locator: "#late"}},
		{"ASSERT_VISIBLE text", &AssertVisible{Locator: `"Nighttroop"`}},
		{"ASSERT_TEXT_VISIBLE", &AssertTextVisible{Text: "Nighttroop"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			page, err := browser.NewPage()
			require.NoError(t, err)
			require.NoError(t, page.SetContent(latePage))
			require.NoError(t, test.assert.Execute(t, map[string]any{}, page, nil))
		})
	}
}

// The page moves to /next half a second after it loads, as one does after a click goes to the next page
const movingPage = `<html><body><script>
setTimeout(() => { history.pushState({}, '', '/next') }, 500)
</script></body></html>`

func TestPathAssertWaitsForThePage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(movingPage))
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
	_, err = page.Goto(server.URL + "/first")
	require.NoError(t, err)
	require.NoError(t, (&AssertPath{URLPath: "/next"}).Execute(t, map[string]any{}, page, nil))
}
