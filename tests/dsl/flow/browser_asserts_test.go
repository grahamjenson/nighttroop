package flow

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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

// The first "Settings" is in a closed menu, and the second is shown, as on a page with a settings menu, and a toast
// covers the button for half a second, as one does after a save
const hiddenFirstPage = `<html><body>
<nav hidden><a>Settings</a></nav>
<h1>Settings</h1>
<button id="save" style="position: absolute; top: 0">Save</button>
<div id="toast" style="position: absolute; top: 0; width: 200px; height: 50px; background: white">Saved</div>
<script>setTimeout(() => document.getElementById("toast").remove(), 500)</script>
</body></html>`

// A visible match is found as soon as it is shown, without waiting for the hidden one before it, and a covered element
// is waited for until it is uncovered
func TestVisibleAssertsWaitOnlyAsLongAsNeeded(t *testing.T) {
	pw, err := playwright.Run()
	require.NoError(t, err)
	t.Cleanup(func() { _ = pw.Stop() })
	browser, err := pw.Chromium.Launch()
	require.NoError(t, err)
	t.Cleanup(func() { _ = browser.Close() })

	page, err := browser.NewPage()
	require.NoError(t, err)
	require.NoError(t, page.SetContent(hiddenFirstPage))
	start := time.Now()
	require.NoError(t, (&AssertTextVisible{Text: "Settings"}).Execute(t, map[string]any{}, page, nil))
	require.Less(t, time.Since(start), time.Second)
	require.NoError(t, (&AssertVisible{Locator: "#save"}).Execute(t, map[string]any{}, page, nil))
}

// Two rows match until half a second after the page loads, when they are swapped out, as a filter's HTMX swap removes
// the rows it leaves out
const swappedRowsPage = `<html><body><table id="events"><tr><td>Site created</td></tr><tr><td>Site created</td></tr></table>
<script>setTimeout(() => { document.getElementById("events").innerHTML = '<tr><td>Sync turned off</td></tr>' }, 500)</script>
</body></html>`

// ASSERT_NOT_VISIBLE waits for every match to go, not refusing a locator that finds several before the swap
func TestNotVisibleWaitsForEveryMatchToGo(t *testing.T) {
	pw, err := playwright.Run()
	require.NoError(t, err)
	t.Cleanup(func() { _ = pw.Stop() })
	browser, err := pw.Chromium.Launch()
	require.NoError(t, err)
	t.Cleanup(func() { _ = browser.Close() })

	page, err := browser.NewPage()
	require.NoError(t, err)
	require.NoError(t, page.SetContent(swappedRowsPage))
	// Both rows are shown before the swap, so the assert finds several, as a strict WaitFor refused at once
	rows, err := page.Locator("#events >> text=Site created").Count()
	require.NoError(t, err)
	require.Equal(t, 2, rows)
	require.NoError(t, (&AssertNotVisible{Locator: "#events >> text=Site created"}).Execute(t, map[string]any{}, page, nil))
	// It passed by waiting for the swap, not by finding nothing before it
	swapped, err := page.Locator("#events >> text=Sync turned off").IsVisible()
	require.NoError(t, err)
	require.True(t, swapped)
}

// The page shows its heading and a select a second and a half after it loads: longer than the 1s default timeout the
// BDD tests set, as a busy CI runner can take to answer even for an element already on the page
const slowPage = `<html><body><script>
setTimeout(() => { document.body.innerHTML = '<h1 id="slow">Nighttroop</h1>' +
	'<select id="pick"><option value="a">A</option><option value="b">B</option></select>' }, 1500)
</script></body></html>`

// An assert or action reading an element waits as long as one finding it, not the browser's 1s default, so a slow
// runner does not fail it
func TestReadsWaitPastTheDefaultTimeout(t *testing.T) {
	pw, err := playwright.Run()
	require.NoError(t, err)
	t.Cleanup(func() { _ = pw.Stop() })
	browser, err := pw.Chromium.Launch()
	require.NoError(t, err)
	t.Cleanup(func() { _ = browser.Close() })

	tests := []struct {
		name string
		run  func(t *testing.T, page playwright.Page) error
	}{
		{"ASSERT_TEXT_EQUALS", func(t *testing.T, page playwright.Page) error {
			return (&AssertTextEquals{Locator: "#slow", Text: "Nighttroop"}).Execute(t, map[string]any{}, page, nil)
		}},
		{"SAVE_TEXT", func(t *testing.T, page playwright.Page) error {
			ctx := map[string]any{}
			_, err := (&SaveText{Locator: "#slow", Key: "heading"}).Execute(t, page, ctx)
			require.Equal(t, "Nighttroop", ctx["heading"])
			return err
		}},
		// A select is picked from rather than typed into, which needs its tag read in time
		{"INPUT a select", func(t *testing.T, page playwright.Page) error {
			if _, err := (&Input{Locator: "#pick", Value: "b"}).Execute(t, page, map[string]any{}); err != nil {
				return err
			}
			value, err := page.Locator("#pick").InputValue()
			require.Equal(t, "b", value)
			return err
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			page, err := browser.NewPage()
			require.NoError(t, err)
			page.SetDefaultTimeout(1000)
			require.NoError(t, page.SetContent(slowPage))
			require.NoError(t, test.run(t, page))
		})
	}
}
