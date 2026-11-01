package bdd_test

import (
	"context"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/background"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/server"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/s3utils"
	"nighttroop.t1d.nz/utils/testutils"
)

// TestSetup runs the setup files against a new server, whose control admin is not made yet
func TestSetup(t *testing.T) {
	runOnOwnServer(t, "setup/first_run.bdd", "192.0.2.10", nil, func(controlDB *db.ControlDB) map[string]any {
		require.NotEmpty(t, controlDB.SetupCode())
		return map[string]any{"setup_code": controlDB.SetupCode()}
	})
}

// TestResourceAlerts runs the resources files against a server whose alert levels are all 0, so every check warns
func TestResourceAlerts(t *testing.T) {
	alwaysHigh := func(c *config.ServerConfig) {
		c.MemoryAlertPercent, c.DiskAlertPercent, c.CPUPeggedAlertSeconds = 0, 0, 0
	}
	runOnOwnServer(t, "resources/alerts.bdd", "192.0.2.11", alwaysHigh, func(controlDB *db.ControlDB) map[string]any {
		require.NoError(t, controlDB.SeedDevAdmin(t.Context()))
		return nil
	})
}

// TestTestSiteBanner runs the testsite files against a server set up as a test site, whose every page says so
func TestTestSiteBanner(t *testing.T) {
	testSite := func(c *config.ServerConfig) { c.TestSite = true }
	runOnOwnServer(t, "testsite/banner.bdd", "192.0.2.12", testSite, func(controlDB *db.ControlDB) map[string]any {
		require.NoError(t, controlDB.SeedDevAdmin(t.Context()))
		return nil
	})
}

// runOnOwnServer runs one file against a server and databases of its own, from the test config as configure changes
// it, with the values prepare gives the file. Its browser sends its own address, so its attempts do not spend
// TestSite's login limit for 127.0.0.1, whose forwarding is trusted.
func runOnOwnServer(t *testing.T, path string, address string, configure func(*config.ServerConfig), prepare func(*db.ControlDB) map[string]any) {
	// The test config is loaded once for every test here, with the login burst TestMain sets for TestSite
	pw, err := playwright.Run()
	require.NoError(t, err)
	t.Cleanup(func() { _ = pw.Stop() })
	browser, err := pw.Chromium.Launch()
	require.NoError(t, err)
	t.Cleanup(func() { _ = browser.Close() })

	config := *testutils.GetTestServerConfig()
	if configure != nil {
		configure(&config)
	}
	controlDB, err := db.NewControlDB(context.Background(), &config)
	require.NoError(t, err)
	values := prepare(controlDB)
	backgroundJobs := background.NewBackGroundManager(context.Background(), controlDB, &config, s3utils.NewS3Client(&config))

	ts := httptest.NewServer(server.GetMainHandler(controlDB, &config))
	t.Cleanup(ts.Close)
	u, err := url.Parse(ts.URL)
	require.NoError(t, err)
	config.HostName = "localhost"
	config.HostPort, err = strconv.Atoi(u.Port())
	require.NoError(t, err)

	browserContext, err := browser.NewContext(playwright.BrowserNewContextOptions{Locale: playwright.String("en"), TimezoneId: playwright.String("UTC"),
		ExtraHttpHeaders: map[string]string{"X-Forwarded-For": address}})
	require.NoError(t, err)
	browserContext.SetDefaultTimeout(1000)
	actionContext := map[string]any{
		"server_url":  ts.URL,
		"control_url": strings.Replace(ts.URL, "127.0.0.1", "control.localhost", 1),
	}
	for key, value := range values {
		actionContext[key] = value
	}
	root, _, _ := strings.Cut(path, "/")
	RunTheActionFileAsATest(t, root, path, browserContext, backgroundJobs, actionContext)
	browserContext.Close()
}
