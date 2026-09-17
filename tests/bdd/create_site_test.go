package bdd_test

import (
	"fmt"
	"net/http/httptest"
	"testing"

	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/tests/bdd"
)

func TestCreateSite(t *testing.T) {

	// Create the router directly
	router := bdd.NewRouter(t)

	// Start an HTTP server around the router
	server := httptest.NewServer(router)
	t.Cleanup(server.Close)

	// Start Playwright
	pw, err := playwright.Run()
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = pw.Stop()
	})

	// Start browser
	browser, err := pw.Chromium.Launch()
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = browser.Close()
	})

	context, err := browser.NewContext()
	require.NoError(t, err)

	context.SetDefaultTimeout(100) // 1 second

	page, err := context.NewPage()
	require.NoError(t, err)

	// Navigate to your Go server
	resp, err := page.Goto(server.URL + "/")

	require.NoError(t, err)
	require.NotNil(t, resp)
	require.Equal(t, 200, resp.Status())

	html, err := page.Content()
	require.NoError(t, err)

	fmt.Println(html)
	// Assertions
	innerText, err := page.Locator("#create-site-button").InnerText()
	require.NoError(t, err)

	require.Equal(t, "Get Started", innerText)
}
