package flow

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/mxschmitt/playwright-go"
)

// ScreenshotDir is where screenshots are kept, each in a folder of its kind: the repository's .test from a test
// package's folder, or the folder a script run is given
var ScreenshotDir = filepath.Join("..", "..", ".test")

// ScreenshotScale is the device pixels a test's browser draws each CSS pixel with, so screenshots are 3/4 the size of
// the page, which keeps their text readable in a file about half the size
const ScreenshotScale = 0.75

func ScreenshotPage(page playwright.Page, foldername string, testPath string, name string) (string, error) {
	root := filepath.Join(ScreenshotDir, foldername)

	dir := filepath.Join(root, strings.TrimSuffix(testPath, filepath.Ext(testPath)))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	// Let web fonts finish loading so the image is the same from run to run.
	if _, err := page.Evaluate("document.fonts.ready"); err != nil {
		return "", err
	}
	// A full page screenshot draws the sticky header where the window is scrolled to, part way down the page, so it is
	// taken from the top of the page, with the header at the top, and the page scrolled back after
	scrolled, err := page.Evaluate("() => { const y = window.scrollY; window.scrollTo(0, 0); return y; }")
	if err != nil {
		return "", err
	}
	defer page.Evaluate("y => window.scrollTo(0, y)", scrolled)
	_, err = page.Screenshot(playwright.PageScreenshotOptions{
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

type BrowserAssert interface {
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
	text, err := Render(r.Text, ctx)
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
	// Contained is ASSERT_CONTAINED: as ASSERT_VISIBLE, and wholly inside, side to side, every box it is in, so nothing
	// too wide pushes it past the edge of its card or form, which the page fitting its window does not show
	Contained bool

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

// readWait is the timeout of reading an element an assert or action has found, such as its text: longer than the 1s
// default, which a busy CI runner can take to answer even for an element on the page
var readWait = playwright.Float(visibleWait)

// waitForVisible waits until any element the locator finds is visible, or visibleWait passes. It waits for a visible
// one, not the first, as the first can stay hidden, as a text in a closed menu does, while another is shown.
// It never fails: the assert after it checks, and fails with the page it found.
func waitForVisible(locator playwright.Locator) {
	_ = locator.Locator("visible=true").First().WaitFor(playwright.LocatorWaitForOptions{State: playwright.WaitForSelectorStateVisible, Timeout: playwright.Float(visibleWait)})
}

// clickableScript scrolls the page, as a user would, to bring the element into the window or out from under
// something fixed to it such as the header, but not boxes that scroll inside the page, then checks the element is
// what a click at its centre would hit
const clickableScript = `el => {
	const hits = () => {
		const r = el.getBoundingClientRect();
		const hit = document.elementFromPoint(r.left + r.width / 2, r.top + r.height / 2);
		return hit !== null && (hit === el || el.contains(hit));
	};
	if (hits()) return true;
	const r = el.getBoundingClientRect();
	window.scrollBy(r.left + r.width / 2 - innerWidth / 2, r.top + r.height / 2 - innerHeight / 2);
	return hits();
}`

// insideGraphScript checks the element is drawn wholly inside the SVG it is in, which clips anything past its edges
const insideGraphScript = `el => {
	const graph = el.closest('svg');
	if (graph === null) return true;
	const r = el.getBoundingClientRect(), g = graph.getBoundingClientRect();
	return r.left >= g.left && r.right <= g.right && r.top >= g.top && r.bottom <= g.bottom;
}`

// outsideScript is "" when the element is wholly inside, side to side, each box it is in, or the first box it pokes
// past and by how much
const outsideScript = `el => {
	const r = el.getBoundingClientRect();
	for (let box = el.parentElement; box !== null && box !== document.body; box = box.parentElement) {
		const b = box.getBoundingClientRect();
		const over = Math.max(b.left - r.left, r.right - b.right);
		if (over > 0.5) {
			const name = box.id ? "#" + box.id : box.tagName.toLowerCase() + (box.classList.length ? "." + [...box.classList].join(".") : "");
			return name + " by " + Math.round(over) + "px";
		}
	}
	return "";
}`

func (r *AssertVisible) Execute(t *testing.T, ctx map[string]any, page playwright.Page, resps []playwright.Response) error {
	// Its locator can use the template, e.g. for a time as it is shown now
	rendered, err := Render(r.Locator, ctx)
	if err != nil {
		t.Fatalf("\n\n%s\n failed to render the locator: %v\n\n", r.ToErrorLine(), err)
	}
	locator := getLocator(rendered, page)
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
		inside, err := locator.Evaluate(insideGraphScript, nil, playwright.LocatorEvaluateOptions{Timeout: readWait})
		if err != nil {
			t.Fatalf("\n\n%s\n failed to check locator is inside its graph: %v, error: %v\n\n%s\n\n", r.ToErrorLine(), r.Locator, err, prettyOutput)
		}
		if inside != true {
			t.Fatalf("\n\n%s\n assertDrawn failed: element is cut off by the edge of its graph: %s\n\n%s\n\n", r.ToErrorLine(), r.Locator, prettyOutput)
		}
		return nil
	}
	// It can be covered for a moment, as by a toast or a modal closing, so it is checked until visibleWait passes
	clickable, err := locator.Evaluate(clickableScript, nil, playwright.LocatorEvaluateOptions{Timeout: readWait})
	for deadline := time.Now().Add(visibleWait * time.Millisecond); err == nil && clickable != true && time.Now().Before(deadline); {
		time.Sleep(50 * time.Millisecond)
		clickable, err = locator.Evaluate(clickableScript, nil, playwright.LocatorEvaluateOptions{Timeout: readWait})
	}
	if err != nil {
		t.Fatalf("\n\n%s\n failed to check locator can be clicked: %v, error: %v\n\n%s\n\n", r.ToErrorLine(), r.Locator, err, prettyOutput)
	}
	if clickable != true {
		t.Fatalf("\n\n%s\n assertVisible failed: element is clipped or covered, so a user cannot see or click it: %s\n\n%s\n\n", r.ToErrorLine(), r.Locator, prettyOutput)
	}
	if r.Contained {
		outside, err := locator.Evaluate(outsideScript, nil, playwright.LocatorEvaluateOptions{Timeout: readWait})
		if err != nil {
			t.Fatalf("\n\n%s\n failed to check locator is inside the boxes it is in: %v, error: %v\n\n%s\n\n", r.ToErrorLine(), r.Locator, err, prettyOutput)
		}
		if outside != "" {
			t.Fatalf("\n\n%s\n assertContained failed: %s pokes past the edge of %v\n\n%s\n\n", r.ToErrorLine(), r.Locator, outside, prettyOutput)
		}
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

// Execute waits up to visibleWait for every element the locator finds to be hidden or gone, as an HTMX swap can take
// a moment. Several can match until the swap, such as the rows a filter is about to remove, which a strict WaitFor
// refused at once rather than waiting for.
func (r *AssertNotVisible) Execute(t *testing.T, ctx map[string]any, page playwright.Page, resps []playwright.Response) error {
	locator := getLocator(r.Locator, page)
	if locator == nil {
		t.Fatalf("\n\n%s\n getLocator returned nil for rawLocator: %v\n\n", r.ToErrorLine(), r.Locator)
	}
	for deadline := time.Now().Add(visibleWait * time.Millisecond); ; time.Sleep(50 * time.Millisecond) {
		visible, err := anyVisible(locator)
		if err == nil && !visible {
			return nil
		}
		if time.Now().After(deadline) {
			html, _ := page.Content()
			t.Fatalf("\n\n%s\n assertNotVisible failed: still visible after %dms (%v)\n\n%s\n\n", r.ToErrorLine(), visibleWait, err, PrettifyHTML(html))
		}
	}
}

// anyVisible is whether any element the locator finds is visible
func anyVisible(locator playwright.Locator) (bool, error) {
	all, err := locator.All()
	if err != nil {
		return false, err
	}
	for _, element := range all {
		if visible, err := element.IsVisible(); err != nil || visible {
			return visible, err
		}
	}
	return false, nil
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

	expected := r.Text
	locator := getLocator(r.Locator, page)
	if locator == nil {
		t.Fatalf("\n\n%s\n getLocator returned nil for rawLocator: %v\n\n%s\n\n", r.ToErrorLine(), r.Locator, prettyOutput)
	}
	innerText, err := locator.InnerText(playwright.LocatorInnerTextOptions{Timeout: readWait})
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

type BrowserAssertStatus struct {
	Code int

	TestPath string
	TestLine int
}

// statusResourceTypes are the requests whose status ASSERT_STATUS checks: a page's and htmx's, not the scripts,
// styles, images and fonts they load, which can come back 304 from the browser's cache, or still be loading from the
// page before when the action starts
var statusResourceTypes = map[string]bool{"document": true, "fetch": true, "xhr": true}

func (r *BrowserAssertStatus) Execute(t *testing.T, ctx map[string]any, page playwright.Page, resps []playwright.Response) error {
	var checked []playwright.Response
	for _, resp := range resps {
		if statusResourceTypes[resp.Request().ResourceType()] {
			checked = append(checked, resp)
		}
	}
	if len(checked) == 0 {
		t.Fatalf("\n\n%s\n assertStatus failed: no page or htmx response\n\n", r.ToErrorLine())
	}
	for _, resp := range checked {
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

func (r *BrowserAssertStatus) ToErrorLine() string {
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

// Execute waits up to visibleWait for the page's path, as after a click the browser may still be going to the next
// page. Its {{ }} values are filled in, e.g. a site ID saved earlier.
func (r *AssertPath) Execute(t *testing.T, ctx map[string]any, page playwright.Page, resps []playwright.Response) error {
	want, err := Render(r.URLPath, ctx)
	if err != nil {
		t.Fatalf("%s: %v", r.ToErrorLine(), err)
	}
	r = &AssertPath{URLPath: want, TestPath: r.TestPath, TestLine: r.TestLine}
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

// AssertCount is how many elements the locator finds are visible
type AssertCount struct {
	Locator string
	Count   int

	TestPath string
	TestLine int
}

// Execute counts the visible elements once, without waiting, as what it counts, such as toasts, can time out
func (r *AssertCount) Execute(t *testing.T, ctx map[string]any, page playwright.Page, resps []playwright.Response) error {
	locator := getLocator(r.Locator, page)
	if locator == nil {
		t.Fatalf("\n\n%s\n getLocator returned nil for rawLocator: %v\n\n", r.ToErrorLine(), r.Locator)
	}
	all, err := locator.All()
	if err != nil {
		t.Fatalf("\n\n%s\n failed to find elements: %v\n\n", r.ToErrorLine(), err)
	}
	count := 0
	for _, element := range all {
		if visible, err := element.IsVisible(); err == nil && visible {
			count++
		}
	}
	if count != r.Count {
		html, _ := page.Content()
		t.Fatalf("\n\n%s\n assertCount failed: found %d visible\n\n%s\n\n", r.ToErrorLine(), count, PrettifyHTML(html))
	}
	return nil
}

func (r *AssertCount) ToErrorLine() string {
	return fmt.Sprintf("%s:%d: expected %d visible elements for locator %s", r.TestPath, r.TestLine, r.Count, r.Locator)
}

// AssertTitle checks the browser tab's title, e.g. ASSERT_TITLE Glucose Readings · Nighttroop
// AssertFits is the page fitting in the window, wide and high, so none of it needs scrolling to see
type AssertFits struct {
	TestPath string
	TestLine int
}

// fitsScript is "" when the page fits in the window, or how far it is past it. The page cannot scroll to what is
// pushed past its left edge, as by a row of buttons pushed to the right, so that is cut off, unless a box it is in
// scrolls or clips it
const fitsScript = `() => {
	const d = document.documentElement;
	if (d.scrollWidth > innerWidth || d.scrollHeight > innerHeight) {
		return "the page is " + d.scrollWidth + "x" + d.scrollHeight + " in a " + innerWidth + "x" + innerHeight + " window";
	}
	const cut = [...document.body.querySelectorAll("*")].find(el => {
		const r = el.getBoundingClientRect();
		if (r.width === 0 || r.left > -1 || !el.checkVisibility({visibilityProperty: true, opacityProperty: true})) return false;
		for (let box = el.parentElement; box !== null; box = box.parentElement) {
			if (getComputedStyle(box).overflowX !== "visible") return false;
		}
		return true;
	});
	return cut === undefined ? "" : (cut.id ? "#" + cut.id : cut.tagName) + " is cut off " + Math.round(-cut.getBoundingClientRect().left) + "px past the left of the window";
}`

func (r *AssertFits) Execute(t *testing.T, ctx map[string]any, page playwright.Page, resps []playwright.Response) error {
	over, err := page.Evaluate(fitsScript)
	if err != nil {
		t.Fatalf("\n\n%s\n failed to measure the page: %v\n\n", r.ToErrorLine(), err)
	}
	if over != "" {
		t.Fatalf("\n\n%s\n assertFits failed: %v\n\n", r.ToErrorLine(), over)
	}
	return nil
}

func (r *AssertFits) ToErrorLine() string {
	return fmt.Sprintf("%s:%d: expected the page to fit in its window", r.TestPath, r.TestLine)
}

type AssertTitle struct {
	Title string

	TestPath string
	TestLine int
}

// Execute waits up to visibleWait for the title, as htmx sets it once a page opened from a menu is swapped in
func (r *AssertTitle) Execute(t *testing.T, ctx map[string]any, page playwright.Page, resps []playwright.Response) error {
	title := ""
	for deadline := time.Now().Add(visibleWait * time.Millisecond); ; time.Sleep(50 * time.Millisecond) {
		var err error
		if title, err = page.Title(); err != nil {
			t.Fatalf("\n\n%s\n failed to read the page's title: %v\n\n", r.ToErrorLine(), err)
		}
		if title == r.Title || time.Now().After(deadline) {
			break
		}
	}
	if title != r.Title {
		t.Fatalf("\n\n%s\n assertTitle failed: expected title %q, got %q\n\n", r.ToErrorLine(), r.Title, title)
	}
	return nil
}

func (r *AssertTitle) ToErrorLine() string {
	return fmt.Sprintf("%s:%d: expected title to be %s", r.TestPath, r.TestLine, r.Title)
}

// //
// ASSERT NO PAGE ERRORS
// //

// AssertNoPageErrors checks the page has thrown no uncaught error since the flow went to it, such as a promise
// rejected with nothing to catch it, which a person only sees in the browser's console
type AssertNoPageErrors struct {
	TestPath string
	TestLine int
}

// pageErrorSettle is how long the assert waits for an error a step just caused, as a rejected promise is reported a
// moment after it
const pageErrorSettle = 300 * time.Millisecond

func (r *AssertNoPageErrors) Execute(t *testing.T, ctx map[string]any, page playwright.Page, resps []playwright.Response) error {
	time.Sleep(pageErrorSettle)
	errors, _ := ctx[pageErrorsKey].(*pageErrors)
	if errors == nil {
		return nil
	}
	errors.Lock()
	thrown := slices.Clone(errors.thrown[page])
	errors.Unlock()
	if len(thrown) > 0 {
		t.Fatalf("\n\n%s\n assertNoPageErrors failed: the page threw:\n  %s\n\n", r.ToErrorLine(), strings.Join(thrown, "\n  "))
	}
	return nil
}

func (r *AssertNoPageErrors) ToErrorLine() string {
	return fmt.Sprintf("%s:%d: expected the page to throw no uncaught error", r.TestPath, r.TestLine)
}
