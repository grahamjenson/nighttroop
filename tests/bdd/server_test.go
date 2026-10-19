package bdd_test

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/mxschmitt/playwright-go"
	"nighttroop.t1d.nz/background"
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

func RunTheActionFileAsATest(t *testing.T, root string, path string, context playwright.BrowserContext, serverURL string, backgroundJobs *background.BackGroundManager, siteInformation *db.SiteCreateResult, defaultJwt string) {
	page, err := context.NewPage()
	if err != nil {
		t.Fatalf("failed to create new page: %v", err)
	}

	removeNightscout := strings.TrimPrefix(path, root+"/")
	t.Run(removeNightscout, func(t *testing.T) {
		// Always keep a screenshot of the page a failing file stopped on
		t.Cleanup(func() {
			if t.Failed() {
				if screenshotPath, err := browser.ScreenshotPage(page, "allScreenshots", path, "failed"); err == nil {
					t.Logf("Failure screenshot saved to: %s", screenshotPath)
				}
			}
		})

		actionFileBytes, err := embeddedFiles.ReadFile(path)
		if err != nil {
			t.Fatalf("failed to read action file: %v", err)
		}
		fmt.Println()
		fmt.Println("===============================================")
		fmt.Printf("Running action file: %s\n", path)
		fmt.Println("===============================================")

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
			fmt.Println()
			fmt.Println("=============== " + action.ToFlatString() + "  ================")
			fmt.Println()
			action.PrintAction()
			fmt.Println()
			fmt.Println("- - - - - - - - - - - - - - - - ")
			fmt.Println()
			// default information

			page, err = action.Execute(t, page, actionContext)
			if err != nil {
				t.Fatalf("failed to execute action: %v, action: %s", err, action.ToFlatString())
			}
			if page == nil {
				t.Fatalf("page is nil after executing action: %s", action.ToFlatString())
			}
			backgroundJobs.RunAllNow(t.Context())
			// A screenshot of every step is slow, so only take them with go test -v
			if testing.Verbose() {
				screenshotPath, err := browser.ScreenshotPage(page, "allScreenshots", path, fmt.Sprintf("%03d", i))
				if err != nil {
					t.Fatalf("failed to take screenshot: %v", err)
				}
				fmt.Println("Screenshot saved to:", screenshotPath)
			}
			fmt.Println("===============================")
		}

	})
}
