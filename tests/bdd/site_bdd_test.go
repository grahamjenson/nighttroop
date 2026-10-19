package bdd_test

import (
	"context"
	"embed"
	"io/fs"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/require"
	"nighttroop.t1d.nz/background"
	"nighttroop.t1d.nz/server"
	"nighttroop.t1d.nz/utils/consts"
	"nighttroop.t1d.nz/utils/s3utils"
	"nighttroop.t1d.nz/utils/testutils"
)

//go:embed site/**
var embeddedFiles embed.FS

func TestSite(t *testing.T) {
	// Every scenario logs in from 127.0.0.1, so the per-IP login burst is shared across all files
	t.Setenv("SITE_LOGIN_PER_MINUTE_BURST", "100")

	config, controlDB, siteInformation, err := testutils.GetConfigControlDBAndSiteDBForTest()

	if err != nil {
		t.Fatalf("failed to get config and databases: %v", err)
	}

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

	testutils.DumpOnFailure(t, controlDB, "test_site_bdd")

	err = fs.WalkDir(embeddedFiles, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		if filepath.Ext(path) != ".bdd" {
			return nil
		}

		context, err := browser.NewContext()

		require.NoError(t, err)

		context.SetDefaultTimeout(1000) // 1 second
		// Accept confirm dialogs, so tests can click buttons with hx-confirm
		context.OnDialog(func(dialog playwright.Dialog) { _ = dialog.Accept() })

		RunTheActionFileAsATest(t, "site", path, context, server.URL, backgroundJobs, siteInformation, defaultJwt)
		context.Close()
		return nil
	})

	if err != nil {
		t.Fatalf("failed walking embedded files: %v", err)
	}
}
