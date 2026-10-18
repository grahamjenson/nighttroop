package bdd_test

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/mxschmitt/playwright-go"
	"nighttroop.t1d.nz/data/db"
	"nighttroop.t1d.nz/server"
	"nighttroop.t1d.nz/tests/dsl/browser"
	"nighttroop.t1d.nz/utils/testutils"
)

func NewRouter(t *testing.T) http.Handler {
	config, controlDB, _, err := testutils.GetConfigControlDBAndSiteDBForTest()

	if err != nil {
		t.Fatalf("failed to get config and databases: %v", err)
	}

	router := server.GetMainHandler(controlDB, config)
	return router
}

func RunTheActionFileAsATest(t *testing.T, root string, path string, context playwright.BrowserContext, serverURL string, siteInformation *db.SiteCreateResult, defaultJwt string) {
	page, err := context.NewPage()
	if err != nil {
		t.Fatalf("failed to create new page: %v", err)
	}

	removeNightscout := strings.TrimPrefix(path, root+"/")
	t.Run(removeNightscout, func(t *testing.T) {
		actionFileBytes, err := embeddedFiles.ReadFile(path)
		if err != nil {
			t.Fatalf("failed to read action file: %v", err)
		}

		actionContext := map[string]any{
			"server_url":    serverURL,
			"default_site":  siteInformation.SiteDB.Subdomain + ".ns.localhost",
			"default_token": siteInformation.PwdNSToken,
			"default_jwt":   defaultJwt,
		}
		actionFile, err := browser.ParseActionFile(string(actionFileBytes), actionContext, path)
		if err != nil {
			t.Fatalf("failed to parse action file: %v", err)
		}
		for i, action := range actionFile.Actions {
			// t.Logf keeps each trace with its subtest, so go test only prints the failing ones
			t.Logf("=============== %s ================", action.ToFlatString())
			// default information

			page, err = action.Execute(t, page, actionContext)
			if err != nil {
				t.Fatalf("failed to execute action: %v, action: %s", err, action.ToFlatString())
			}
			if page == nil {
				t.Fatalf("page is nil after executing action: %s", action.ToFlatString())
			}
			screenshotPath, err := browser.ScreenshotPage(page, "allScreenshots", path, fmt.Sprintf("%03d", i))
			if err != nil {
				t.Fatalf("failed to take screenshot: %v", err)
			}
			t.Logf("Screenshot saved to: %s", screenshotPath)
		}

	})
}
