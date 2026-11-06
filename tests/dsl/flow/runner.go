package flow

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/mxschmitt/playwright-go"
	"nighttroop.t1d.nz/tests/dsl/browser"
	"nighttroop.t1d.nz/tests/dsl/socketio"
)

// Runner runs a flow's steps in order, all with the same context, so each sees what those before it got
type Runner struct {
	// Client sends the HTTP steps' requests, as jsonapi.Client does to a test server, whatever their URL's host
	Client *http.Client
	// RunBackgroundJobs runs every background job, for RUN_BACKGROUND_JOBS
	RunBackgroundJobs func()
	// NewPage opens the page for the flow's first browser step
	NewPage func() (playwright.Page, error)
	// Page is the page the browser steps are on, nil until the first opens it
	Page playwright.Page
}

// Run runs the flow, failing the test at the first step that cannot finish. A socket stays connected through the
// browser steps between its API steps.
func (r *Runner) Run(t *testing.T, flow *Flow, ctx map[string]any) {
	t.Helper()
	api := &socketio.Runner{Client: r.Client, RunBackgroundJobs: r.RunBackgroundJobs}
	defer api.Close()
	for i, step := range flow.Steps {
		if step.API != nil {
			api.Step(t, flow.API, step.API, ctx)
			continue
		}
		r.browse(t, flow.Path, i, step.Browser, ctx)
	}
}

// browse runs one browser action, on the flow's page
func (r *Runner) browse(t *testing.T, path string, i int, action browser.Action, ctx map[string]any) {
	t.Helper()
	if r.Page == nil {
		page, err := r.NewPage()
		if err != nil {
			t.Fatalf("failed to create new page: %v", err)
		}
		r.Page = page
	}
	fmt.Printf("\n=============== %s  ================\n\n", action.ToFlatString())
	action.PrintAction()
	fmt.Println("\n- - - - - - - - - - - - - - - - ")
	if _, ok := action.(*browser.RunBackgroundJobs); ok {
		r.RunBackgroundJobs()
	}
	// A failing action answers no page, so Page keeps the one it failed on, for a failure screenshot
	next, err := action.Execute(t, r.Page, ctx)
	if err != nil {
		t.Fatalf("failed to execute action: %v, action: %s", err, action.ToFlatString())
	}
	if next == nil {
		t.Fatalf("page is nil after executing action: %s", action.ToFlatString())
	}
	r.Page = next
	// A screenshot of every step is slow, so only take them with go test -v
	if testing.Verbose() {
		screenshotPath, err := browser.ScreenshotPage(r.Page, "allScreenshots", path, fmt.Sprintf("%03d", i))
		if err != nil {
			t.Fatalf("failed to take screenshot: %v", err)
		}
		fmt.Println("Screenshot saved to:", screenshotPath)
	}
	fmt.Println("===============================")
}
