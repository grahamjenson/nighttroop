package bdd_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/mxschmitt/playwright-go"
	"nighttroop.t1d.nz/background"
	"nighttroop.t1d.nz/tests/dsl/browser"
)

// RunTheActionFileAsATest runs the action file at path, its templates filled from actionContext
func RunTheActionFileAsATest(t *testing.T, root string, path string, context playwright.BrowserContext, backgroundJobs *background.BackGroundManager, actionContext map[string]any) {
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

		fmt.Println()
		fmt.Println("===============================================")
		fmt.Printf("Running action file: %s\n", path)
		fmt.Println("===============================================")

		actionFile, err := browser.ParseActionFile(embeddedFiles, path)
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

			if _, ok := action.(*browser.RunBackgroundJobs); ok {
				backgroundJobs.RunAllNow(t.Context())
			}
			page, err = action.Execute(t, page, actionContext)
			if err != nil {
				t.Fatalf("failed to execute action: %v, action: %s", err, action.ToFlatString())
			}
			if page == nil {
				t.Fatalf("page is nil after executing action: %s", action.ToFlatString())
			}
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
