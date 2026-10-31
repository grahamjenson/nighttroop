package bdd_test

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/background"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/server"
	"nighttroop.t1d.nz/utils/config"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/s3utils"
	"nighttroop.t1d.nz/utils/testutils"
)

//go:embed site/** control/** steps/** setup/** resources/**
var embeddedFiles embed.FS

// bddShards is how many groups of scenarios run at once. Each has its own databases, server and background jobs, so
// one group's background jobs never touch another's sites; only the browser is shared. A scenario's test is
// TestSite/shardN/<file>, so run one with -run 'TestSite//summary'.
const bddShards = 4

func TestSite(t *testing.T) {
	// Every scenario logs in from 127.0.0.1, so the per-IP login burst is shared across all files
	t.Setenv("SITE_LOGIN_PER_MINUTE_BURST", "100")

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

	// Loaded once before the shards start, as loading parses the command line's flags
	testConfig := testutils.GetTestServerConfig()

	paths := []string{}
	err = fs.WalkDir(embeddedFiles, ".", func(path string, d fs.DirEntry, err error) error {
		// The setup and resources files need servers of their own, see TestSetup and TestResourceAlerts
		ownServer := strings.HasPrefix(path, "setup/") || strings.HasPrefix(path, "resources/")
		if err == nil && !d.IsDir() && filepath.Ext(path) == ".bdd" && !ownServer {
			paths = append(paths, path)
		}
		return err
	})
	if err != nil {
		t.Fatalf("failed walking embedded files: %v", err)
	}

	for shard := range bddShards {
		t.Run(fmt.Sprintf("shard%d", shard), func(t *testing.T) {
			t.Parallel()
			shardPaths := []string{}
			for i := shard; i < len(paths); i += bddShards {
				shardPaths = append(shardPaths, paths[i])
			}
			// Each has its own copy, as it sets its own server's port
			config := *testConfig
			runScenarios(t, browser, &config, shardPaths)
		})
	}
}

// runScenarios runs each scenario file in a new browser context against a server and databases of their own
func runScenarios(t *testing.T, browser playwright.Browser, config *config.ServerConfig, paths []string) {
	controlDB, err := db.NewControlDB(context.Background(), config)
	require.NoError(t, err)
	require.NoError(t, controlDB.SeedDevAdmin())
	siteInformation, err := controlDB.CreateSite()
	require.NoError(t, err)

	defaultJwt, _, err := config.Security.GenerateUserJWTToken(siteInformation.SiteDB.Subdomain, consts.FIRST_PWD_NAME, string(consts.IdentityRoleDevicePWDPhone))
	if err != nil {
		t.Fatalf("failed to generate JWT token: %v", err)
	}

	require.NoError(t, err)
	require.NotEmpty(t, defaultJwt) // TODO insert this as a cookie to dodge login for default

	router := server.GetMainHandler(controlDB, config)
	// Background jobs are run after every action instead of on their timers
	backgroundJobs := background.NewBackGroundManager(context.Background(), controlDB, config, s3utils.NewS3Client(config))

	server := httptest.NewServer(router)
	t.Cleanup(server.Close)

	u, err := url.Parse(server.URL)
	if err != nil {
		t.Fatal(err)
	}

	port, err := strconv.Atoi(u.Port())
	if err != nil {
		t.Fatal(err)
	}
	config.HostName = "localhost"
	config.HostPort = port

	testutils.DumpOnFailure(t, controlDB, "test_site_bdd")

	for _, path := range paths {
		// Sites take the time zone and language of the browser that creates them, and en names no country, so a
		// test's site keeps the default units and formats
		context, err := browser.NewContext(playwright.BrowserNewContextOptions{Locale: playwright.String("en"), TimezoneId: playwright.String("UTC")})

		require.NoError(t, err)

		context.SetDefaultTimeout(1000) // 1 second
		// Accept confirm dialogs, so tests can click buttons with hx-confirm
		context.OnDialog(func(dialog playwright.Dialog) { _ = dialog.Accept() })

		RunTheActionFileAsATest(t, "site", path, context, backgroundJobs, map[string]any{
			"server_url":    server.URL,
			"control_url":   strings.Replace(server.URL, "127.0.0.1", "control.localhost", 1),
			"default_site":  siteInformation.SiteDB.Subdomain + ".ns.localhost",
			"default_token": siteInformation.PwdNSToken,
			"default_jwt":   defaultJwt,
		})
		context.Close()
	}
}
