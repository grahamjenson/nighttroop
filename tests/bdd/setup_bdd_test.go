package bdd_test

import (
	"context"
	"fmt"
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
	"nighttroop.t1d.nz/tests/dsl/flow"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/s3utils"
	"nighttroop.t1d.nz/utils/testutils"
)

// ownServer is a file that needs a server and databases of its own, from the test config as configure changes it,
// with the values prepare gives the file. Its browser sends its own address, so its attempts do not spend TestSite's
// login limit for 127.0.0.1, whose forwarding is trusted.
type ownServer struct {
	path      string
	address   string
	configure func(*config.ServerConfig)
	prepare   func(*testing.T, *db.ControlDB) map[string]any
}

// ownServers are every file TestOwnServer runs, which TestSite leaves out
var ownServers = []ownServer{
	// A new server, whose control admin is not made yet
	{"setup/first_run.bdd", "192.0.2.10", nil, func(t *testing.T, controlDB *db.ControlDB) map[string]any {
		require.NotEmpty(t, controlDB.SetupCode())
		return map[string]any{"setup_code": controlDB.SetupCode()}
	}},
	// A server whose alert levels are all 0, and disk warns however much is free, so every check warns
	{"resources/alerts.bdd", "192.0.2.11", func(c *config.ServerConfig) {
		c.MemoryAlertPercent, c.DiskAlertPercent, c.CPUPeggedAlertSeconds = 0, 0, 0
		c.DiskAlertFreeGB = 1 << 30
	}, seedDevAdmin},
	// A server set up as a test site, whose every page says so
	{"testsite/banner.bdd", "192.0.2.12", func(c *config.ServerConfig) { c.TestSite = true }, seedDevAdmin},
	// A server allowing 10 sites, 9 of which it has
	{"sitelimit/banner.bdd", "192.0.2.13", func(c *config.ServerConfig) { c.MaxSites = 10 }, func(t *testing.T, controlDB *db.ControlDB) map[string]any {
		seedDevAdmin(t, controlDB)
		for range 9 {
			_, err := controlDB.CreateSite(t.Context())
			require.NoError(t, err)
		}
		return nil
	}},
	// A server with a feedback link, on its home page and every footer
	{"feedback/links.bdd", "192.0.2.14", func(c *config.ServerConfig) { c.FeedbackURL = "https://github.com/grahamjenson/nighttroop/issues" }, seedDevAdmin},
}

func seedDevAdmin(t *testing.T, controlDB *db.ControlDB) map[string]any {
	require.NoError(t, controlDB.SeedDevAdmin(t.Context()))
	return nil
}

// TestOwnServer runs each of the ownServers, one at a time, on one browser. A file's test is
// TestOwnServer/<folder>/<file>, so run one with -run 'TestOwnServer/setup'.
func TestOwnServer(t *testing.T) {
	pw, err := playwright.Run()
	require.NoError(t, err)
	t.Cleanup(func() { _ = pw.Stop() })
	browser, err := pw.Chromium.Launch()
	require.NoError(t, err)
	t.Cleanup(func() { _ = browser.Close() })

	for _, own := range ownServers {
		root, _, _ := strings.Cut(own.path, "/")
		t.Run(root, func(t *testing.T) { runOnOwnServer(t, browser, root, own) })
	}
}

// runOnOwnServer runs one of the ownServers' files in a new browser context, against a server of its own
func runOnOwnServer(t *testing.T, browser playwright.Browser, root string, own ownServer) {
	// The test config is loaded once for every test here, with the login burst TestMain sets for TestSite
	config := *testutils.GetTestServerConfig()
	if own.configure != nil {
		own.configure(&config)
	}
	controlDB, err := db.NewControlDB(context.Background(), &config)
	require.NoError(t, err)
	values := own.prepare(t, controlDB)
	backgroundJobs := background.NewBackGroundManager(context.Background(), controlDB, &config, s3utils.NewS3Client(&config))

	ts := httptest.NewServer(server.WithCrossOriginProtection(server.GetMainHandler(controlDB, &config)))
	t.Cleanup(ts.Close)
	u, err := url.Parse(ts.URL)
	require.NoError(t, err)
	config.HostName = "localhost"
	config.HostPort, err = strconv.Atoi(u.Port())
	require.NoError(t, err)

	browserContext, err := browser.NewContext(playwright.BrowserNewContextOptions{Locale: playwright.String("en"), TimezoneId: playwright.String("UTC"),
		DeviceScaleFactor: playwright.Float(flow.ScreenshotScale),
		ExtraHttpHeaders:  map[string]string{"X-Forwarded-For": own.address}})
	require.NoError(t, err)
	browserContext.SetDefaultTimeout(1000)
	actionContext := map[string]any{
		"server_url":  ts.URL,
		"scheme":      "http",
		"host":        fmt.Sprintf("localhost:%d", config.HostPort),
		"control_url": strings.Replace(ts.URL, "127.0.0.1", "control.localhost", 1),
	}
	for key, value := range values {
		actionContext[key] = value
	}
	RunTheActionFileAsATest(t, root, own.path, ts, browserContext, backgroundJobs, actionContext)
	browserContext.Close()
}
