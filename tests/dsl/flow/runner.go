package flow

import (
	"fmt"
	"net/http"
	"os"
	"sync"
	"testing"

	"github.com/mxschmitt/playwright-go"
)

// StepScreenshots is whether a screenshot is taken after every browser step, into allScreenshots, which about doubles
// a browser test's time. It is set by FLOW_STEP_SCREENSHOTS, rather than go test -v, as -json, gotestsum and editors
// run tests verbosely too.
var StepScreenshots = os.Getenv("FLOW_STEP_SCREENSHOTS") != ""

// Runner runs a flow's steps in order, all with the same context, so each sees what those before it got
type Runner struct {
	// Client sends the HTTP steps' requests, as Client does to a test server, whatever their URL's host
	Client *http.Client
	// RunBackgroundJobs runs every background job, for RUN_BACKGROUND_JOBS
	RunBackgroundJobs func()
	// NewSocket makes a Socket.IO client speaking that Engine.IO version, for CONNECT
	NewSocket func(eio int) Socket
	// NewPage opens the page for the flow's first browser step
	NewPage func() (playwright.Page, error)
	// Page is the page the browser steps are on, nil until the first opens it
	Page playwright.Page
	// errors are the uncaught errors each page the flow went to threw, which ASSERT_NO_PAGE_ERRORS checks
	errors *pageErrors
}

// pageErrorsKey is where the flow's context keeps its pageErrors, for ASSERT_NO_PAGE_ERRORS
const pageErrorsKey = "__page_errors"

// pageErrors are the uncaught errors, including promises rejected with nothing to catch them, each page threw since the
// runner first went to it, as the browser's console shows them
type pageErrors struct {
	sync.Mutex
	thrown map[playwright.Page][]string
}

// watch records the uncaught errors page throws from now on, once for each page
func (r *Runner) watch(page playwright.Page, ctx map[string]any) {
	if r.errors == nil {
		r.errors = &pageErrors{thrown: map[playwright.Page][]string{}}
	}
	ctx[pageErrorsKey] = r.errors
	r.errors.Lock()
	defer r.errors.Unlock()
	if _, ok := r.errors.thrown[page]; ok {
		return
	}
	r.errors.thrown[page] = []string{}
	page.OnPageError(func(err error) {
		r.errors.Lock()
		defer r.errors.Unlock()
		r.errors.thrown[page] = append(r.errors.thrown[page], err.Error())
	})
}

// Run runs the flow, failing the test at the first step that cannot finish. A socket stays connected through the
// browser steps between its API steps.
func (r *Runner) Run(t *testing.T, flow *Flow, ctx map[string]any) {
	t.Helper()
	api := &httpRunner{Client: r.Client, RunBackgroundJobs: r.RunBackgroundJobs, NewSocket: r.NewSocket}
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
func (r *Runner) browse(t *testing.T, path string, i int, action Action, ctx map[string]any) {
	t.Helper()
	if r.Page == nil {
		page, err := r.NewPage()
		if err != nil {
			t.Fatalf("failed to create new page: %v", err)
		}
		r.Page = page
	}
	r.watch(r.Page, ctx)
	fmt.Printf("\n=============== %s  ================\n\n", action.ToFlatString())
	action.PrintAction()
	fmt.Println("\n- - - - - - - - - - - - - - - - ")
	if _, ok := action.(*RunBackgroundJobs); ok {
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
	r.watch(r.Page, ctx)
	if StepScreenshots {
		screenshotPath, err := ScreenshotPage(r.Page, "allScreenshots", path, fmt.Sprintf("%03d", i))
		if err != nil {
			t.Fatalf("failed to take screenshot: %v", err)
		}
		fmt.Println("Screenshot saved to:", screenshotPath)
	}
	fmt.Println("===============================")
}
