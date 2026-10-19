package browser

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

type ActionFile struct {
	Name    string
	Actions []Action
}

//////////////////
// ACTIONS
//////////////////

// /////
// GOTO
// //////
type Goto struct {
	URL string

	Asserts []Assert

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

func (r *Goto) AddAssert(assert Assert) {
	r.Asserts = append(r.Asserts, assert)
}

func (r *Goto) Execute(t *testing.T, page playwright.Page, ctx map[string]any) (playwright.Page, error) {

	url, err := renderTemplate(r.URL, ctx)
	if err != nil {
		t.Fatalf(
			"\n\n GOTO FAILED \n %s: %v \n\n",
			r.ToErrorLine("Error Rendering URL"),
			err,
		)
	}

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
// CLICK
//////

type Click struct {
	Locator string

	Asserts []Assert

	TestPath string
	TestLine int
}

func (r *Click) ToFlatString() string {
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

func (r *Click) AddAssert(assert Assert) {
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
		_, ok := a.(*AssertStatus)
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

	locator := getLocator(r.Locator, page)
	if locator == nil {
		t.Fatalf("Failed to get locator for %s\n\n%s\n\n", r.Locator, prettyOutput)
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
	err = locator.Click()
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
		case *AssertStatus:
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

	err = locator.Clear()
	if err != nil {
		return nil, err
	}
	// Use this instead of Fill to simulate typing each key sequentially
	err = locator.PressSequentially(r.Value)
	if err != nil {
		return nil, err
	}

	// Implement the execution logic for the Input action here
	return page, nil
}

// ///
// SET
// ///
type Set struct {
	Key   string
	Value string

	TestPath string
	TestLine int
}

func (r *Set) ToFlatString() string {
	return fmt.Sprintf("SET %s %s", r.Key, r.Value)
}

func (r *Set) ToErrorLine(msg string) string {
	if msg == "" {
		return fmt.Sprintf("%s:%d: %s", r.TestPath, r.TestLine, r.ToFlatString())
	}
	return fmt.Sprintf("%s:%d: %s: %s", r.TestPath, r.TestLine, msg, r.ToFlatString())
}

func (r *Set) PrintAction() {
	fmt.Println(r.ToFlatString())
}

func (r *Set) Execute(t *testing.T, page playwright.Page, ctx map[string]any) (playwright.Page, error) {
	// Implement the execution logic for the Set action here
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
	value, err := renderTemplate(r.Value, ctx)
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
