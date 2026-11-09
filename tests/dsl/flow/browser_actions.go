package flow

import (
	"fmt"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/mxschmitt/playwright-go"
)

type Action interface {
	ToErrorLine(msg string) string
	ToFlatString() string
	PrintAction()
	Execute(t *testing.T, page playwright.Page, ctx map[string]any) (playwright.Page, error)
}

//////////////////
// ACTIONS
//////////////////

// /////
// GOTO
// //////
type Goto struct {
	URL string

	Asserts []BrowserAssert

	TestPath string
	TestLine int
}

func (r *Goto) ToFlatString() string {
	return fmt.Sprintf("GOTO %s", r.URL)
}

func (r *Goto) ToErrorLine(msg string) string {
	if msg == "" {
		return fmt.Sprintf("%s:%d: %s", r.TestPath, r.TestLine, r.ToFlatString())
	}
	return fmt.Sprintf("%s:%d: %s: %s", r.TestPath, r.TestLine, msg, r.ToFlatString())
}

func (r *Goto) PrintAction() {
	fmt.Println(r.ToFlatString())
}

func (r *Goto) AddAssert(assert BrowserAssert) {
	r.Asserts = append(r.Asserts, assert)
}

func (r *Goto) Execute(t *testing.T, page playwright.Page, ctx map[string]any) (playwright.Page, error) {

	url, err := Render(r.URL, ctx)
	if err != nil {
		t.Fatalf(
			"\n\n GOTO FAILED \n %s: %v \n\n",
			r.ToErrorLine("Error Rendering URL"),
			err,
		)
	}

	// A path is relative to the current page, so a file can move around the site it logged in to
	url = resolveAgainstPage(page, url)

	resp, err := page.Goto(url)
	if err != nil {
		t.Fatalf(
			"\n\n GOTO FAILED \n %s: %v \n URL: %v \n\n",
			r.ToErrorLine("Error with GOTO"),
			err,
			url,
		)
	}

	if resp == nil {
		t.Fatalf(
			"\n\n GOTO FAILED \n %s: response is nil\n\n",
			r.ToErrorLine("Error with GOTO"),
		)
	}

	for _, assert := range r.Asserts {
		assert.Execute(t, ctx, page, []playwright.Response{resp})
	}

	// Implement the execution logic for the Goto action here
	return page, nil
}

//////
// POST
//////

// Post sends URL its JSON Body from the page, with the page's cookies, failing the test unless it succeeds
type Post struct {
	URL  string
	Body string

	TestPath string
	TestLine int
}

func (r *Post) ToFlatString() string {
	return fmt.Sprintf("POST %s %s", r.URL, r.Body)
}

func (r *Post) ToErrorLine(msg string) string {
	if msg == "" {
		return fmt.Sprintf("%s:%d: %s", r.TestPath, r.TestLine, r.ToFlatString())
	}
	return fmt.Sprintf("%s:%d: %s: %s", r.TestPath, r.TestLine, msg, r.ToFlatString())
}

func (r *Post) PrintAction() {
	fmt.Println(r.ToFlatString())
}

func (r *Post) Execute(t *testing.T, page playwright.Page, ctx map[string]any) (playwright.Page, error) {
	url, err := Render(r.URL, ctx)
	if err != nil {
		t.Fatalf("%s: %v", r.ToErrorLine("Error Rendering URL"), err)
	}
	body, err := Render(r.Body, ctx)
	if err != nil {
		t.Fatalf("%s: %v", r.ToErrorLine("Error Rendering Body"), err)
	}
	status, err := page.Evaluate(`async ([url, body]) => (await fetch(url, {method: "POST", body,
		headers: {"Content-Type": "application/json", "Accept": "application/json"}})).status`, []any{resolveAgainstPage(page, url), body})
	if err != nil {
		t.Fatalf("%s: %v", r.ToErrorLine("Error with POST"), err)
	}
	if code, ok := status.(int); !ok || code >= 400 {
		t.Fatalf("%s: status %v", r.ToErrorLine("POST failed"), status)
	}
	return page, nil
}

//////
// CLICK
//////

type Click struct {
	Locator string
	// History is "back" or "forward" for CLICK_BACK and CLICK_FORWARD, the browser's buttons, instead of clicking Locator
	History string
	// Key is pressed instead of clicking, for PRESS
	Key string
	// Text is typed into Locator, a key at a time, instead of clicking it, for TYPE
	Text string
	// Hidden hides the page or shows it again, as a browser does a tab in the background, for BACKGROUND and FOREGROUND
	Hidden *bool
	// Corner clicks just inside Locator's top left corner instead of its centre, for CLICK_CORNER,
	// e.g. a modal's backdrop beside the dialog over its centre
	Corner bool
	// Scroll turns the mouse wheel by this many pixels, down or up when negative, as a user scrolls the page, for SCROLL
	Scroll *int
	// Quickly are CSS selectors clicked one straight after another, not waiting for each to finish, for CLICK_QUICKLY
	Quickly []string
	// Asks leaves the site's confirm dialog open, for CLICK_ASKS, to check what it asks; a plain CLICK confirms it, as a
	// person going on with what they clicked does
	Asks bool
	// Hover moves the mouse over Locator instead of clicking it, this share of the way across it and half way down, for
	// HOVER, e.g. to point at a time on a graph
	Hover *float64

	Asserts []BrowserAssert

	TestPath string
	TestLine int
}

func (r *Click) ToFlatString() string {
	if r.History != "" {
		return "CLICK_" + strings.ToUpper(r.History)
	}
	if r.Key != "" {
		return "PRESS " + r.Key
	}
	if r.Corner {
		return "CLICK_CORNER " + r.Locator
	}
	if r.Scroll != nil {
		return fmt.Sprintf("SCROLL %d", *r.Scroll)
	}
	if r.Hover != nil {
		return fmt.Sprintf("HOVER %g%% %s", *r.Hover*100, r.Locator)
	}
	if len(r.Quickly) > 0 {
		return "CLICK_QUICKLY " + strings.Join(r.Quickly, ", ")
	}
	if r.Asks {
		return "CLICK_ASKS " + r.Locator
	}
	return fmt.Sprintf("CLICK %s", r.Locator)
}

func (r *Click) ToErrorLine(msg string) string {
	if msg == "" {
		return fmt.Sprintf("%s:%d: %s", r.TestPath, r.TestLine, r.ToFlatString())
	}
	return fmt.Sprintf("%s:%d: %s: %s", r.TestPath, r.TestLine, msg, r.ToFlatString())
}

func (r *Click) PrintAction() {
	fmt.Println(r.ToFlatString())
}

func (r *Click) AddAssert(assert BrowserAssert) {
	r.Asserts = append(r.Asserts, assert)
}

func (r *Click) NeedsResponseOrPage() (bool, bool) {
	// check for new page first
	if len(r.Asserts) == 0 {
		return false, false
	}

	for _, a := range r.Asserts {
		// if there is a AssertNewPage assert, we need a new page
		_, ok := a.(*AssertNewPage)
		if ok {
			return false, true
		}
	}

	for _, a := range r.Asserts {
		// if there is a AssertResponse assert, we need a response
		_, ok := a.(*BrowserAssertStatus)
		if ok {
			return true, false
		}
	}

	return false, false
}

func (r *Click) Execute(t *testing.T, page playwright.Page, ctx map[string]any) (playwright.Page, error) {
	context := page.Context()

	html, err := page.Content()
	if err != nil {
		return nil, err
	}
	prettyOutput := PrettifyHTML(html)

	var locator playwright.Locator
	if r.History == "" && r.Key == "" && r.Hidden == nil && r.Scroll == nil && len(r.Quickly) == 0 {
		if locator = getLocator(r.Locator, page); locator == nil {
			t.Fatalf("Failed to get locator for %s\n\n%s\n\n", r.Locator, prettyOutput)
		}
	}

	var responses []playwright.Response
	var requests []playwright.Request
	var newPages []playwright.Page

	context.On("page", func(p playwright.Page) {
		newPages = append(newPages, p)
	})

	context.On("request", func(req playwright.Request) {
		requests = append(requests, req)
	})

	context.On("response", func(resp playwright.Response) {
		responses = append(responses, resp)
	})

	oldURL := page.URL()
	switch {
	case r.History == "back":
		_, err = page.GoBack()
	case r.History == "forward":
		_, err = page.GoForward()
	case r.Key != "":
		err = page.Keyboard().Press(r.Key)
	case r.Hidden != nil:
		_, err = page.Evaluate(`hidden => {
			Object.defineProperty(document, "hidden", {configurable: true, get: () => hidden});
			Object.defineProperty(document, "visibilityState", {configurable: true, get: () => hidden ? "hidden" : "visible"});
			document.dispatchEvent(new Event("visibilitychange"));
		}`, *r.Hidden)
	case r.Scroll != nil:
		err = page.Mouse().Wheel(0, float64(*r.Scroll))
	case len(r.Quickly) > 0:
		// Clicked in the page in one go, as quickly as a request can be sent, so each replaces the one before
		_, err = page.Evaluate(`selectors => selectors.forEach(selector => {
			const element = document.querySelector(selector);
			if (!element) throw new Error("CLICK_QUICKLY found no " + selector);
			element.click();
		})`, r.Quickly)
	case r.Text != "":
		// TYPE's text is filled in like INPUT's value, as a TOTP code made from a site's secret is
		var text string
		if text, err = Render(r.Text, ctx); err == nil {
			err = locator.PressSequentially(text, playwright.LocatorPressSequentiallyOptions{Timeout: playwright.Float(actionWait)})
		}
	case r.Corner:
		err = locator.Click(playwright.LocatorClickOptions{Timeout: playwright.Float(actionWait), Position: &playwright.Position{X: 2, Y: 2}})
	case r.Hover != nil:
		var box *playwright.Rect
		if box, err = locator.BoundingBox(playwright.LocatorBoundingBoxOptions{Timeout: playwright.Float(actionWait)}); err == nil {
			err = locator.Hover(playwright.LocatorHoverOptions{Timeout: playwright.Float(actionWait), Position: &playwright.Position{X: box.Width * *r.Hover, Y: box.Height / 2}})
		}
	default:
		err = clickAndAnswer(page, locator, r.Asks)
	}
	if err != nil {
		t.Fatalf(
			"\n\n CLICK FAILED \n %s: %v \n\n %s \n\n",
			r.ToErrorLine("Error with CLICK"),
			prettyOutput,
			err,
		)
	}

	// We have to wait for the reponses to be captured. An assert that needs a
	// response, a new page or a new URL keeps waiting (bounded) until it
	// arrives: on a busy CI runner this can take well over 50ms.
	time.Sleep(50 * time.Millisecond)
	needsResponse, needsPage, needsURL := false, false, false
	for _, a := range r.Asserts {
		switch a.(type) {
		case *BrowserAssertStatus:
			needsResponse = true
		case *AssertNewPage:
			needsPage = true
		case *AssertNewURL:
			needsURL = true
		}
	}
	for deadline := time.Now().Add(5 * time.Second); time.Now().Before(deadline) &&
		((needsResponse && len(responses) == 0) || (needsPage && len(newPages) == 0) || (needsURL && page.URL() == oldURL)); {
		time.Sleep(20 * time.Millisecond)
	}

	newURL := false
	if page.URL() != oldURL {
		newURL = true
	}

	if len(newPages) > 1 {
		t.Fatalf(
			"\n\n CLICK FAILED \n %s: multiple new pages captured\n\n %s \n\n",
			r.ToErrorLine("Multiple new pages with CLICK"),
			prettyOutput,
		)
	}

	newPage := first(newPages)
	if newPage != nil {
		page = newPage

		if strings.Contains(page.URL(), "chrome-error") || page.URL() == "about:blank#blocked" {
			t.Fatalf("\n\n%s\n new page failed to load properly\nURL returnes as %s, might be bad URL\n\n", r.ToErrorLine("Error with NEW PAGE"), page.URL())
		}
		err := page.WaitForLoadState(
			playwright.PageWaitForLoadStateOptions{
				State: playwright.LoadStateDomcontentloaded,
			},
		)
		if err != nil {
			return nil, err
		}
	}

	for _, assert := range r.Asserts {
		if _, ok := assert.(*AssertNewPage); ok {
			if newPage == nil {
				t.Fatalf("\n\n%s\n assertNewPage failed: no new page was opened\n\n", r.ToErrorLine("Error with ASSERT"))
			}
			continue
		}

		if _, ok := assert.(*AssertNewURL); ok {
			if !newURL {
				t.Fatalf("\n\n%s\n assertNewURL failed: no new URL was opened\n old URL: %s\n new URL: %s\n\n", r.ToErrorLine("Error with ASSERT"), oldURL, page.URL())
			}
			continue
		}
		err := assert.Execute(t, ctx, page, responses)
		if err != nil {
			t.Fatalf(
				"\n\n ASSERT FAILED \n %s: %v \n\n %s \n\n",
				r.ToErrorLine("Error with ASSERT"),
				err,
				prettyOutput,
			)
		}
	}

	return page, nil
}

func first[T any](slice []T) T {
	var zero T
	if len(slice) == 0 {
		return zero
	}
	return slice[0]
}

/////
// INPUT
//////

type Input struct {
	Locator string
	Value   string

	TestPath string
	TestLine int
}

func (r *Input) ToFlatString() string {
	return fmt.Sprintf("INPUT %s %s", r.Locator, r.Value)
}

func (r *Input) ToErrorLine(msg string) string {
	if msg == "" {
		return fmt.Sprintf("%s:%d: %s", r.TestPath, r.TestLine, r.ToFlatString())
	}
	return fmt.Sprintf("%s:%d: %s: %s", r.TestPath, r.TestLine, msg, r.ToFlatString())
}

func (r *Input) PrintAction() {
	fmt.Println(r.ToFlatString())
}

func (r *Input) Execute(t *testing.T, page playwright.Page, ctx map[string]any) (playwright.Page, error) {
	html, err := page.Content()
	if err != nil {
		return nil, err
	}
	prettyOutput := PrettifyHTML(html)

	locator := getLocator(r.Locator, page)
	if locator == nil {
		t.Fatalf("Failed to get locator for %s\n\n%s\n\n", r.Locator, prettyOutput)
	}

	// The value can be one saved earlier, e.g. INPUT #key {{ .backup_key }}
	value, err := Render(r.Value, ctx)
	if err != nil {
		t.Fatalf("\n\n %s: %v \n\n", r.ToErrorLine("Error Rendering value"), err)
	}

	// A select can't be typed into, so pick the option with this value, or else this label, as when its value is an ID
	if tag, err := locator.Evaluate("element => element.tagName", nil, playwright.LocatorEvaluateOptions{Timeout: readWait}); err == nil && tag == "SELECT" {
		_, err = locator.SelectOption(playwright.SelectOptionValues{ValuesOrLabels: &[]string{value}}, playwright.LocatorSelectOptionOptions{Timeout: playwright.Float(actionWait)})
		return page, err
	}

	// A date or time input splits typing across its parts, so it takes the value whole, e.g. 2026-01-02T12:30, as does a
	// text area, which is pasted into, such as with a key
	if kind, err := locator.Evaluate("element => element.type", nil, playwright.LocatorEvaluateOptions{Timeout: readWait}); err == nil && (kind == "datetime-local" || kind == "date" || kind == "time" || kind == "textarea") {
		return page, locator.Fill(value, playwright.LocatorFillOptions{Timeout: playwright.Float(actionWait)})
	}

	err = locator.Clear(playwright.LocatorClearOptions{Timeout: playwright.Float(actionWait)})
	if err != nil {
		return nil, err
	}
	// Typed a key at a time, not filled, as a person types. actionWait is for the input to be ready: the keys are typed
	// once it has focus, with no deadline, as a busy runner can take longer than actionWait to type a long value
	if err = locator.Focus(playwright.LocatorFocusOptions{Timeout: playwright.Float(actionWait)}); err != nil {
		return nil, err
	}
	if err = page.Keyboard().Type(value); err != nil {
		return nil, err
	}

	// Implement the execution logic for the Input action here
	return page, nil
}

// ///
// COOKIES
// ///

// SaveCookie stores the value of a cookie for the current page in the context under Key
type SaveCookie struct {
	Name string
	Key  string

	TestPath string
	TestLine int
}

func (r *SaveCookie) ToFlatString() string {
	return fmt.Sprintf("SAVE_COOKIE %s %s", r.Name, r.Key)
}

func (r *SaveCookie) ToErrorLine(msg string) string {
	return fmt.Sprintf("%s:%d: %s: %s", r.TestPath, r.TestLine, msg, r.ToFlatString())
}

func (r *SaveCookie) PrintAction() {
	fmt.Println(r.ToFlatString())
}

func (r *SaveCookie) Execute(t *testing.T, page playwright.Page, ctx map[string]any) (playwright.Page, error) {
	cookies, err := page.Context().Cookies(page.URL())
	if err != nil {
		return nil, err
	}
	for _, c := range cookies {
		if c.Name == r.Name {
			ctx[r.Key] = c.Value
			return page, nil
		}
	}
	t.Fatalf("\n\n %s \n\n", r.ToErrorLine("cookie not found"))
	return page, nil
}

// SetCookie replaces a cookie for the current page's host, e.g. to replay one taken from another site
type SetCookie struct {
	Name  string
	Value string

	TestPath string
	TestLine int
}

func (r *SetCookie) ToFlatString() string {
	return fmt.Sprintf("SET_COOKIE %s %s", r.Name, r.Value)
}

func (r *SetCookie) ToErrorLine(msg string) string {
	return fmt.Sprintf("%s:%d: %s: %s", r.TestPath, r.TestLine, msg, r.ToFlatString())
}

func (r *SetCookie) PrintAction() {
	fmt.Println(r.ToFlatString())
}

func (r *SetCookie) Execute(t *testing.T, page playwright.Page, ctx map[string]any) (playwright.Page, error) {
	value, err := Render(r.Value, ctx)
	if err != nil {
		t.Fatalf("\n\n %s: %v \n\n", r.ToErrorLine("Error Rendering value"), err)
	}
	pageURL, err := url.Parse(page.URL())
	if err != nil {
		return nil, err
	}
	// Clear first so the new cookie replaces the server's rather than sitting beside it
	if err := page.Context().ClearCookies(playwright.BrowserContextClearCookiesOptions{Name: r.Name}); err != nil {
		return nil, err
	}
	err = page.Context().AddCookies([]playwright.OptionalCookie{{
		Name:   r.Name,
		Value:  value,
		Domain: playwright.String(pageURL.Hostname()),
		Path:   playwright.String("/"),
	}})
	return page, err
}

// ClearCookie removes a cookie from the browser
type ClearCookie struct {
	Name string

	TestPath string
	TestLine int
}

func (r *ClearCookie) ToFlatString() string {
	return fmt.Sprintf("CLEAR_COOKIE %s", r.Name)
}

func (r *ClearCookie) ToErrorLine(msg string) string {
	return fmt.Sprintf("%s:%d: %s: %s", r.TestPath, r.TestLine, msg, r.ToFlatString())
}

func (r *ClearCookie) PrintAction() {
	fmt.Println(r.ToFlatString())
}

func (r *ClearCookie) Execute(t *testing.T, page playwright.Page, ctx map[string]any) (playwright.Page, error) {
	return page, page.Context().ClearCookies(playwright.BrowserContextClearCookiesOptions{Name: r.Name})
}

// RunBackgroundJobs runs every background and cron job once. The test runner does the work, as it owns the jobs.
type RunBackgroundJobs struct {
	Asserts []BrowserAssert

	TestPath string
	TestLine int
}

func (r *RunBackgroundJobs) AddAssert(assert BrowserAssert) {
	r.Asserts = append(r.Asserts, assert)
}

func (r *RunBackgroundJobs) ToFlatString() string {
	return "RUN_BACKGROUND_JOBS"
}

func (r *RunBackgroundJobs) ToErrorLine(msg string) string {
	return fmt.Sprintf("%s:%d: %s: %s", r.TestPath, r.TestLine, msg, r.ToFlatString())
}

func (r *RunBackgroundJobs) PrintAction() {
	fmt.Println(r.ToFlatString())
}

// Execute runs its asserts; the jobs themselves are run by the test before it, against the test's server
func (r *RunBackgroundJobs) Execute(t *testing.T, page playwright.Page, ctx map[string]any) (playwright.Page, error) {
	for _, assert := range r.Asserts {
		if err := assert.Execute(t, ctx, page, nil); err != nil {
			t.Fatalf("\n\n ASSERT FAILED \n %s: %v \n\n", r.ToErrorLine("Error with ASSERT"), err)
		}
	}
	return page, nil
}

// resolveAgainstPage turns a path like /data/glucose into a full URL on the current page's host
func resolveAgainstPage(page playwright.Page, target string) string {
	current, err := url.Parse(page.URL())
	if err != nil || !strings.HasPrefix(target, "/") {
		return target
	}
	resolved, err := current.Parse(target)
	if err != nil {
		return target
	}
	return resolved.String()
}

// Window sets the page's window size
type Window struct {
	Width, Height int

	TestPath string
	TestLine int
}

func (r *Window) ToFlatString() string {
	return fmt.Sprintf("WINDOW %dx%d", r.Width, r.Height)
}

func (r *Window) ToErrorLine(msg string) string {
	return fmt.Sprintf("%s:%d: %s: %s", r.TestPath, r.TestLine, msg, r.ToFlatString())
}

func (r *Window) PrintAction() {
	fmt.Println(r.ToFlatString())
}

func (r *Window) Execute(t *testing.T, page playwright.Page, ctx map[string]any) (playwright.Page, error) {
	return page, page.SetViewportSize(r.Width, r.Height)
}

// UserAgent sends the page's requests with a User-Agent, or the browser's own when it is empty
type UserAgent struct {
	Value string

	TestPath string
	TestLine int
}

func (r *UserAgent) ToFlatString() string {
	return fmt.Sprintf("USER_AGENT %s", r.Value)
}

func (r *UserAgent) ToErrorLine(msg string) string {
	return fmt.Sprintf("%s:%d: %s: %s", r.TestPath, r.TestLine, msg, r.ToFlatString())
}

func (r *UserAgent) PrintAction() {
	fmt.Println(r.ToFlatString())
}

func (r *UserAgent) Execute(t *testing.T, page playwright.Page, ctx map[string]any) (playwright.Page, error) {
	headers := map[string]string{}
	if r.Value != "" {
		headers["User-Agent"] = r.Value
	}
	return page, page.SetExtraHTTPHeaders(headers)
}

// SaveText keeps an element's text under a key, for a later line to use, e.g. a token shown once
type SaveText struct {
	Locator string
	Key     string
	// Value keeps the input's value, for SAVE_VALUE, rather than its text
	Value bool

	TestPath string
	TestLine int
}

func (r *SaveText) ToFlatString() string {
	if r.Value {
		return fmt.Sprintf("SAVE_VALUE %s %s", r.Locator, r.Key)
	}
	return fmt.Sprintf("SAVE_TEXT %s %s", r.Locator, r.Key)
}

func (r *SaveText) ToErrorLine(msg string) string {
	return fmt.Sprintf("%s:%d: %s: %s", r.TestPath, r.TestLine, msg, r.ToFlatString())
}

func (r *SaveText) PrintAction() {
	fmt.Println(r.ToFlatString())
}

func (r *SaveText) Execute(t *testing.T, page playwright.Page, ctx map[string]any) (playwright.Page, error) {
	locator := page.Locator(r.Locator)
	var text string
	var err error
	if r.Value {
		text, err = locator.InputValue(playwright.LocatorInputValueOptions{Timeout: readWait})
	} else {
		text, err = locator.InnerText(playwright.LocatorInnerTextOptions{Timeout: readWait})
	}
	if err != nil {
		return nil, fmt.Errorf("%s", r.ToErrorLine(err.Error()))
	}
	ctx[r.Key] = strings.TrimSpace(text)
	return page, nil
}

// VirtualAuthenticator gives the page Chrome's virtual authenticator, a platform one keeping passkeys that verifies the
// person without asking, so a page can make and sign in with passkeys
type VirtualAuthenticator struct {
	TestPath string
	TestLine int
}

func (r *VirtualAuthenticator) ToFlatString() string {
	return "VIRTUAL_AUTHENTICATOR"
}

func (r *VirtualAuthenticator) ToErrorLine(msg string) string {
	return fmt.Sprintf("%s:%d: %s: %s", r.TestPath, r.TestLine, msg, r.ToFlatString())
}

func (r *VirtualAuthenticator) PrintAction() {
	fmt.Println(r.ToFlatString())
}

func (r *VirtualAuthenticator) Execute(t *testing.T, page playwright.Page, ctx map[string]any) (playwright.Page, error) {
	session, err := page.Context().NewCDPSession(page)
	if err == nil {
		_, err = session.Send("WebAuthn.enable", map[string]any{})
	}
	if err == nil {
		_, err = session.Send("WebAuthn.addVirtualAuthenticator", map[string]any{"options": map[string]any{
			"protocol": "ctap2", "transport": "internal", "hasResidentKey": true, "hasUserVerification": true,
			"isUserVerified": true, "automaticPresenceSimulation": true,
		}})
	}
	if err != nil {
		return nil, fmt.Errorf("%s", r.ToErrorLine(err.Error()))
	}
	return page, nil
}

// clickAndAnswer clicks locator, and when it asks to be confirmed in the site's confirm dialog waits for the dialog,
// then confirms it unless asks, which leaves it open to be checked
func clickAndAnswer(page playwright.Page, locator playwright.Locator, asks bool) error {
	confirms, err := locator.Evaluate(`element => !!element.closest("[hx-confirm]")`, nil, playwright.LocatorEvaluateOptions{Timeout: playwright.Float(actionWait)})
	if err != nil {
		return err
	}
	if err := locator.Click(playwright.LocatorClickOptions{Timeout: playwright.Float(actionWait)}); err != nil || confirms != true {
		if asks && err == nil {
			return fmt.Errorf("CLICK_ASKS clicked something that asks no confirmation")
		}
		return err
	}
	dialog := page.Locator("dialog#confirm[open]")
	if err := dialog.WaitFor(playwright.LocatorWaitForOptions{Timeout: playwright.Float(actionWait)}); err != nil || asks {
		return err
	}
	return page.Locator("#confirm--ok").Click(playwright.LocatorClickOptions{Timeout: playwright.Float(actionWait)})
}
