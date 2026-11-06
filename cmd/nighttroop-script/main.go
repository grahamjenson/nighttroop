// Command nighttroop-script runs test scripts, .flow files and .http, .sio and .bdd ones, against a Nighttroop server
// running elsewhere, such as staging: to make sites with real data and check its API, sockets, live updates and pages.
//
//	go run ./cmd/nighttroop-script --host staging.nighttroop.nz --root scripts/staging new_site.flow
//
// Each script runs as a test, so -test.v prints each step and -test.run picks the scripts to run. A URL on localhost,
// or on a host under it such as abcdefgh.ns.localhost, goes to that name under --host. {{ server_url }},
// {{ control_url }} and {{ site_url "abcdefgh" }} are the server's addresses, --var KEY=VALUE sets {{ .KEY }}, and
// {{ env "KEY" }} reads an environment variable, so a password need not be written in a script. A script with
// RUN_BACKGROUND_JOBS is refused, as only a test's own server can run its jobs at once. Browser steps need Playwright's
// Chromium: go run github.com/mxschmitt/playwright-go/cmd/playwright install chromium
package main

import (
	"flag"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/mxschmitt/playwright-go"
	"nighttroop.t1d.nz/tests/dsl/browser"
	"nighttroop.t1d.nz/tests/dsl/flow"
	"nighttroop.t1d.nz/tests/dsl/jsonapi"
)

// settings is how scripts are run, from the command line
type settings struct {
	scheme, host string
	vars         map[string]string
	timeout      time.Duration
	headed       bool
}

func main() {
	s := settings{vars: map[string]string{}}
	flag.StringVar(&s.host, "host", "", "the server's host name, with its port if not the scheme's, e.g. staging.nighttroop.nz")
	flag.StringVar(&s.scheme, "scheme", "https", "http or https")
	root := flag.String("root", ".", "the folder the scripts, and the files they INCLUDE, are in")
	flag.StringVar(&browser.ScreenshotDir, "out", "script-screenshots", "the folder for screenshots, SCREENSHOT's and a failed script's")
	flag.DurationVar(&s.timeout, "timeout", 10*time.Second, "how long a browser step waits for the page")
	flag.BoolVar(&s.headed, "headed", false, "show the browser as it runs")
	flag.Func("var", "KEY=VALUE, set as {{ .KEY }} in every script; may be given more than once", func(kv string) error {
		key, value, ok := strings.Cut(kv, "=")
		if !ok || key == "" {
			return fmt.Errorf("%q is not KEY=VALUE", kv)
		}
		s.vars[key] = value
		return nil
	})
	testing.Init()
	flag.Parse()
	if s.host == "" || flag.NArg() == 0 {
		fmt.Fprintln(os.Stderr, "usage: nighttroop-script --host HOST [--root FOLDER] [--var KEY=VALUE]... SCRIPT...")
		flag.PrintDefaults()
		os.Exit(2)
	}
	fsys, client := os.DirFS(*root), jsonapi.RemoteClient(s.scheme, s.host)
	tests := []testing.InternalTest{}
	for _, path := range flag.Args() {
		tests = append(tests, testing.InternalTest{Name: path, F: func(t *testing.T) { run(t, fsys, path, client, s) }})
	}
	testing.Main(regexp.MatchString, tests, nil, nil)
}

// run runs the script at path in fsys, sending its requests with client, and its browser steps, if it has any, in a
// browser of its own
func run(t *testing.T, fsys fs.FS, path string, client *http.Client, s settings) {
	script, err := flow.Parse(fsys, path)
	if err != nil {
		t.Fatalf("failed to parse: %v", err)
	}
	if err := refuseBackgroundJobs(script); err != nil {
		t.Fatal(err)
	}
	ctx := map[string]any{
		"server_url":  s.scheme + "://" + s.host,
		"control_url": s.scheme + "://control." + s.host,
		"scheme":      s.scheme,
		"host":        s.host,
	}
	for key, value := range s.vars {
		ctx[key] = value
	}
	runner := &flow.Runner{Client: client, NewPage: func() (playwright.Page, error) { return newPage(t, s) }}
	t.Cleanup(func() {
		if t.Failed() && runner.Page != nil {
			if screenshot, err := browser.ScreenshotPage(runner.Page, "failed", path, "failed"); err == nil {
				t.Logf("Failure screenshot saved to: %s", screenshot)
			}
		}
	})
	runner.Run(t, script, ctx)
}

// refuseBackgroundJobs refuses a script with RUN_BACKGROUND_JOBS, which only a test's own server can run at once
func refuseBackgroundJobs(script *flow.Flow) error {
	for _, step := range script.Steps {
		_, browsing := step.Browser.(*browser.RunBackgroundJobs)
		if browsing || (step.API != nil && step.API.Verb == "RUN_BACKGROUND_JOBS") {
			return fmt.Errorf("%s: RUN_BACKGROUND_JOBS runs a test server's jobs at once, so it cannot run against another server", script.Path)
		}
	}
	return nil
}

// newPage starts Chromium for a script's first browser step, closed when the script ends
func newPage(t *testing.T, s settings) (playwright.Page, error) {
	pw, err := playwright.Run()
	if err != nil {
		return nil, fmt.Errorf("cannot start Playwright, is its Chromium installed? %w", err)
	}
	t.Cleanup(func() { _ = pw.Stop() })
	chromium, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{Headless: playwright.Bool(!s.headed)})
	if err != nil {
		return nil, err
	}
	t.Cleanup(func() { _ = chromium.Close() })
	context, err := chromium.NewContext()
	if err != nil {
		return nil, err
	}
	context.SetDefaultTimeout(float64(s.timeout.Milliseconds()))
	// Accept confirm dialogs, so a script can click buttons with hx-confirm
	context.OnDialog(func(dialog playwright.Dialog) { _ = dialog.Accept() })
	return context.NewPage()
}
