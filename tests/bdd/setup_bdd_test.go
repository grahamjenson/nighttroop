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
	"nighttroop.t1d.nz/utils/s3utils"
	"nighttroop.t1d.nz/utils/testutils"
)

// TestSetup runs the setup files against a new server, whose control admin is not made yet
func TestSetup(t *testing.T) {
	pw, err := playwright.Run()
	require.NoError(t, err)
	t.Cleanup(func() { _ = pw.Stop() })
	browser, err := pw.Chromium.Launch()
	require.NoError(t, err)
	t.Cleanup(func() { _ = browser.Close() })

	config := *testutils.GetTestServerConfig()
	controlDB, err := db.NewControlDB(context.Background(), &config)
	require.NoError(t, err)
	require.NotEmpty(t, controlDB.SetupCode())
	backgroundJobs := background.NewBackGroundManager(context.Background(), controlDB, &config, s3utils.NewS3Client(&config))

	ts := httptest.NewServer(server.GetMainHandler(controlDB, &config))
	t.Cleanup(ts.Close)
	u, err := url.Parse(ts.URL)
	require.NoError(t, err)
	config.HostName = "localhost"
	config.HostPort, err = strconv.Atoi(u.Port())
	require.NoError(t, err)

	browserContext, err := browser.NewContext(playwright.BrowserNewContextOptions{Locale: playwright.String("en"), TimezoneId: playwright.String("UTC")})
	require.NoError(t, err)
	browserContext.SetDefaultTimeout(1000)
	RunTheActionFileAsATest(t, "setup", "setup/first_run.bdd", browserContext, backgroundJobs, map[string]any{
		"server_url":  ts.URL,
		"control_url": strings.Replace(ts.URL, "127.0.0.1", "control.localhost", 1),
		"setup_code":  controlDB.SetupCode(),
	})
	browserContext.Close()
}
