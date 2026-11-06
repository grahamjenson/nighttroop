package bdd_test

import (
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mxschmitt/playwright-go"
	"nighttroop.t1d.nz/background"
	"nighttroop.t1d.nz/tests/dsl/flow"
)

// RunTheActionFileAsATest runs the .bdd or .flow file at path, its templates filled from actionContext, its API and
// socket steps sent to server
func RunTheActionFileAsATest(t *testing.T, root string, path string, server *httptest.Server, context playwright.BrowserContext, backgroundJobs *background.BackGroundManager, actionContext map[string]any) {
	removeNightscout := strings.TrimPrefix(path, root+"/")
	t.Run(removeNightscout, func(t *testing.T) {
		runner := &flow.Runner{
			Client:            flow.Client(server),
			RunBackgroundJobs: func() { backgroundJobs.RunAllNow(t.Context()) },
			NewPage:           context.NewPage,
		}
		// Always keep a screenshot of the page a failing file stopped on
		t.Cleanup(func() {
			if t.Failed() && runner.Page != nil {
				if screenshotPath, err := flow.ScreenshotPage(runner.Page, "allScreenshots", path, "failed"); err == nil {
					t.Logf("Failure screenshot saved to: %s", screenshotPath)
				}
			}
		})

		fmt.Println()
		fmt.Println("===============================================")
		fmt.Printf("Running action file: %s\n", path)
		fmt.Println("===============================================")

		actionFlow, err := flow.Parse(embeddedFiles, path)
		if err != nil {
			t.Fatalf("failed to parse action file: %v", err)
		}
		runner.Run(t, actionFlow, actionContext)
	})
}
