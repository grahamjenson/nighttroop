package browser

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/mxschmitt/playwright-go"
)

func ScreenshotPage(page playwright.Page, foldername string, testPath string, name string) (string, error) {
	root := filepath.Join("..", "..", ".test", foldername)

	dir := filepath.Join(root, strings.TrimSuffix(testPath, filepath.Ext(testPath)))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	// Let web fonts finish loading so the image is the same from run to run.
	if _, err := page.Evaluate("document.fonts.ready"); err != nil {
		return "", err
	}
	_, err := page.Screenshot(playwright.PageScreenshotOptions{
		Path:       playwright.String(filepath.Join(dir, name+".png")),
		FullPage:   playwright.Bool(true),
		Animations: playwright.ScreenshotAnimationsDisabled,
		Caret:      playwright.ScreenshotCaretHide,
		// A full page screenshot can take longer than the 1s default on a busy CI runner
		Timeout: playwright.Float(screenshotWait),
	})
	return filepath.Join(dir, name+".png"), err
}

// /////////////////////
// ASSERTS
// //////////////////////

type Assert interface {
	Execute(t *testing.T, ctx map[string]any, page playwright.Page, resps []playwright.Response) error
}

// ///
// TAKE A  SCREENSHOT
// ///
type Screenshot struct {
	Name     string
	TestPath string
	TestLine int
}

// Execute writes ./tests/screenshots/<bdd file>/<name>.png we can see the screenshots
func (r *Screenshot) Execute(t *testing.T, ctx map[string]any, page playwright.Page, resps []playwright.Response) error {
	_, err := ScreenshotPage(page, "savedScreenshots", r.TestPath, r.Name)
	return err
}

/////
// ASSERT TEXT VISIBLE
/////

type AssertTextVisible struct {
	Text string

	TestPath string
	TestLine int
}

func (r *AssertTextVisible) Execute(t *testing.T, ctx map[string]any, page playwright.Page, resps []playwright.Response) error {
	text, err := renderTemplate(r.Text, ctx)
	if err != nil {
		t.Fatalf("%s: %v", r.ToErrorLine(), err)
	}
	r = &AssertTextVisible{Text: text, TestPath: r.TestPath, TestLine: r.TestLine}
	waitForVisible(page.GetByText(r.Text, playwright.PageGetByTextOptions{Exact: playwright.Bool(false)}))
	html, err := page.Content()
	if err != nil {
		return err
	}

	prettyOutput := PrettifyHTML(html)

	locators, err := page.GetByText(r.Text, playwright.PageGetByTextOptions{
		Exact: playwright.Bool(false),
	}).All()

	if err != nil {
		t.Fatalf("ASSERT_VISIBLE_TEXT: error getting all locators for %q: %v", r.Text, err)
	}

	if len(locators) == 0 {
		t.Fatalf("\n\n%s\n getLocator returned nil for rawLocator: %v\n\n%s\n\n",
			r.ToErrorLine(), r.Text, prettyOutput)
	}

	anyVisible := false
	for _, locator := range locators {
		visible, err := locator.IsVisible()
		if err != nil {
			t.Fatalf("ASSERT_VISIBLE_TEXT: error checking visibility of %q: %v", r.Text, err)
		}
		if visible {
			anyVisible = true
			break
		}
	}
	if !anyVisible {
		t.Errorf("ASSERT_VISIBLE_TEXT: text %q is not visible", r.Text)
	}
	return nil
}

func (r *AssertTextVisible) ToErrorLine() string {
	return fmt.Sprintf("%s:%d: expected Text '%s' to be visible for user", r.TestPath, r.TestLine, r.Text)
}

// //
// ASSERT VISIBLE
// //

// AssertVisible checks a user can see the element and click it: it is rendered, and it is what a
// click at its centre would hit, so one clipped by a scrolling box or covered by another element fails
type AssertVisible struct {
	Locator string
	// Drawn is ASSERT_DRAWN, for a mark in a graph that others may be drawn over: it is shown whole inside its
	// graph, but need not be what a click at its centre would hit
	Drawn bool

	TestPath string
	TestLine int
}

// visibleWait is how long a visibility assert waits for the page to show what it checks, as after a
// click the browser may still be loading or swapping in the next page
const visibleWait = 5000

// screenshotWait is how long a screenshot may take, in ms
const screenshotWait = 10000

// actionWait is how long a click or input waits for its element to be visible, enabled and stable, in ms:
// longer than the 1s default, which a busy CI runner can take for a transition to settle
const actionWait = visibleWait

// waitForVisible waits until the first element the locator finds is visible, or visibleWait passes.
// It never fails: the assert after it checks, and fails with the page it found.
func waitForVisible(locator playwright.Locator) {
	_ = locator.First().WaitFor(playwright.LocatorWaitForOptions{State: playwright.WaitForSelectorStateVisible, Timeout: playwright.Float(visibleWait)})
}

// clickableScript scrolls the page, as a user would, to bring the element into the window, but not
// boxes that scroll inside the page, then checks the element is what a click at its centre would hit
const clickableScript = `el => {
	let r = el.getBoundingClientRect();
	const x = r.left + r.width / 2, y = r.top + r.height / 2;
	if (x < 0 || x > innerWidth || y < 0 || y > innerHeight) {
		window.scrollBy(x - innerWidth / 2, y - innerHeight / 2);
		r = el.getBoundingClientRect();
	}
	const hit = document.elementFromPoint(r.left + r.width / 2, r.top + r.height / 2);
	return hit !== null && (hit === el || el.contains(hit));
}`

// insideGraphScript checks the element is drawn wholly inside the SVG it is in, which clips anything past its edges
const insideGraphScript = `el => {
	const graph = el.closest('svg');
	if (graph === null) return true;
	const r = el.getBoundingClientRect(), g = graph.getBoundingClientRect();
	return r.left >= g.left && r.right <= g.right && r.top >= g.top && r.bottom <= g.bottom;
}`

func (r *AssertVisible) Execute(t *testing.T, ctx map[string]any, page playwright.Page, resps []playwright.Response) error {
	locator := getLocator(r.Locator, page)
	if locator != nil {
		waitForVisible(locator)
	}
	html, err := page.Content()
	if err != nil {
		return err
	}

	prettyOutput := PrettifyHTML(html)

	if locator == nil {
		t.Fatalf("\n\n%s\n getLocator returned nil for rawLocator: %v\n\n%s\n\n", r.ToErrorLine(), r.Locator, prettyOutput)
	}
	visible, err := locator.IsVisible()
	if err != nil {
		t.Fatalf("\n\n%s\n failed to check visibility for locator: %v, error: %v\n\n%s\n\n", r.ToErrorLine(), r.Locator, err, prettyOutput)
	}
	if !visible {
		t.Fatalf("\n\n%s\n assertVisible failed: expected value to be visible for locator: %s\n\n%s\n\n", r.ToErrorLine(), r.Locator, prettyOutput)
	}
	if r.Drawn {
		inside, err := locator.Evaluate(insideGraphScript, nil)
		if err != nil {
			t.Fatalf("\n\n%s\n failed to check locator is inside its graph: %v, error: %v\n\n%s\n\n", r.ToErrorLine(), r.Locator, err, prettyOutput)
		}
		if inside != true {
			t.Fatalf("\n\n%s\n assertDrawn failed: element is cut off by the edge of its graph: %s\n\n%s\n\n", r.ToErrorLine(), r.Locator, prettyOutput)
		}
		return nil
	}
	clickable, err := locator.Evaluate(clickableScript, nil)
	if err != nil {
		t.Fatalf("\n\n%s\n failed to check locator can be clicked: %v, error: %v\n\n%s\n\n", r.ToErrorLine(), r.Locator, err, prettyOutput)
	}
	if clickable != true {
		t.Fatalf("\n\n%s\n assertVisible failed: element is clipped or covered, so a user cannot see or click it: %s\n\n%s\n\n", r.ToErrorLine(), r.Locator, prettyOutput)
	}
	return nil
}

// //
// ASSERT NOT VISIBLE
// //
type AssertNotVisible struct {
	Locator string

	TestPath string
	TestLine int
}

// Execute waits for the element to be hidden or gone, as an HTMX swap can take a moment
func (r *AssertNotVisible) Execute(t *testing.T, ctx map[string]any, page playwright.Page, resps []playwright.Response) error {
	locator := getLocator(r.Locator, page)
	if locator == nil {
		t.Fatalf("\n\n%s\n getLocator returned nil for rawLocator: %v\n\n", r.ToErrorLine(), r.Locator)
	}
	if err := locator.WaitFor(playwright.LocatorWaitForOptions{State: playwright.WaitForSelectorStateHidden}); err != nil {
		html, _ := page.Content()
		t.Fatalf("\n\n%s\n assertNotVisible failed: %v\n\n%s\n\n", r.ToErrorLine(), err, PrettifyHTML(html))
	}
	return nil
}

func (r *AssertNotVisible) ToErrorLine() string {
	return fmt.Sprintf("%s:%d: expected element not to be visible for locator %s", r.TestPath, r.TestLine, r.Locator)
}

func (r *AssertVisible) ToErrorLine() string {
	return fmt.Sprintf("%s:%d: expected element to be visible for locator %s", r.TestPath, r.TestLine, r.Locator)
}

// ///
// ASSERT TEXT EQUALS
// ///

type AssertTextEquals struct {
	Locator string
	Text    string

	TestPath string
	TestLine int
}

func (r *AssertTextEquals) Execute(t *testing.T, ctx map[string]any, page playwright.Page, resps []playwright.Response) error {
	html, err := page.Content()
	if err != nil {
		return err
	}

	prettyOutput := PrettifyHTML(html)

	expected := normalizeValue(r.Text)
	locator := getLocator(r.Locator, page)
	if locator == nil {
		t.Fatalf("\n\n%s\n getLocator returned nil for rawLocator: %v\n\n%s\n\n", r.ToErrorLine(), r.Locator, prettyOutput)
	}
	innerText, err := locator.InnerText()
	if err != nil {
		t.Fatalf("\n\n%s\n failed to get inner text for locator: %v, error: %v\n\n%s\n\n", r.ToErrorLine(), r.Locator, err, prettyOutput)
	}
	if innerText != expected {
		t.Fatalf("\n\n%s\n pageContains failed: expected %v at %v found %v: \n\n%s\n\n", r.ToErrorLine(), expected, r.Locator, innerText, prettyOutput)
	}
	return nil
}

func (r *AssertTextEquals) ToErrorLine() string {
	return fmt.Sprintf("%s:%d: expected text for locator %s", r.TestPath, r.TestLine, r.Locator)
}

// ////
// ASSERT STATUS
// ////

type AssertStatus struct {
	Code int

	TestPath string
	TestLine int
}

func (r *AssertStatus) Execute(t *testing.T, ctx map[string]any, page playwright.Page, resps []playwright.Response) error {

	if len(resps) == 0 {
		t.Fatalf("\n\n%s\n assertStatus failed: response is nil\n\n", r.ToErrorLine())
	}
	for _, resp := range resps {
		// the body is only read for a failure, as reading a stream's, e.g. an SSE's, waits until it closes
		if resp.Status() != r.Code {
			body := ""
			bodyBytes, err := resp.Body()
			if err != nil {
				body = "Error reading body"
			} else {
				body = string(bodyBytes)
			}

			t.Fatalf("\n\n%s\n assertStatus failed: expected status code %d, got %d for URL %s with Body: %s\n\n", r.ToErrorLine(), r.Code, resp.Status(), resp.URL(), body)
		}
	}

	return nil
}

func (r *AssertStatus) ToErrorLine() string {
	return fmt.Sprintf("%s:%d: expected status code %d", r.TestPath, r.TestLine, r.Code)
}

// ////
// Just checks the new page is there
// ////
type AssertNewPage struct {
	TestPath string
	TestLine int
}

func (r *AssertNewPage) Execute(t *testing.T, ctx map[string]any, page playwright.Page, resps []playwright.Response) error {
	return nil
}

func (r *AssertNewPage) ToErrorLine() string {
	return fmt.Sprintf("%s:%d: expected a new page to be opened", r.TestPath, r.TestLine)
}

type AssertNewURL struct {
	TestPath string
	TestLine int
}

func (r *AssertNewURL) Execute(t *testing.T, ctx map[string]any, page playwright.Page, resps []playwright.Response) error {
	return nil
}

func (r *AssertNewURL) ToErrorLine() string {
	return fmt.Sprintf("%s:%d: expected a new URL to be opened", r.TestPath, r.TestLine)
}

type AssertPath struct {
	URLPath string

	TestPath string
	TestLine int
}

// Execute waits up to visibleWait for the page's path, as after a click the browser may still be going to the next page
func (r *AssertPath) Execute(t *testing.T, ctx map[string]any, page playwright.Page, resps []playwright.Response) error {
	path := ""
	for deadline := time.Now().Add(visibleWait * time.Millisecond); ; time.Sleep(50 * time.Millisecond) {
		parsedURL, err := url.Parse(page.URL())
		if err != nil {
			t.Fatalf("\n\n%s\n failed to parse page URL: %v\n\n", r.ToErrorLine(), err)
		}
		path = parsedURL.Path
		if path == r.URLPath || time.Now().After(deadline) {
			break
		}
	}
	if path != r.URLPath {
		t.Fatalf("\n\n%s\n assertPath failed: expected path %s, got %s\n\n", r.ToErrorLine(), r.URLPath, path)
	}
	return nil
}

func (r *AssertPath) ToErrorLine() string {
	return fmt.Sprintf("%s:%d: expected path to be %s", r.TestPath, r.TestLine, r.URLPath)
}
