package browser

import (
	"fmt"
	"net/url"
	"testing"

	"github.com/mxschmitt/playwright-go"
)

// /////////////////////
// ASSERTS
// //////////////////////

type Assert interface {
	Execute(t *testing.T, ctx map[string]any, page playwright.Page, resps []playwright.Response) error
}

// //
// ASSERT VISIBLE
// //
type AssertVisible struct {
	Locator string

	TestPath string
	TestLine int
}

func (r *AssertVisible) Execute(t *testing.T, ctx map[string]any, page playwright.Page, resps []playwright.Response) error {
	html, err := page.Content()
	if err != nil {
		return err
	}

	prettyOutput := PrettifyHTML(html)

	locator := getLocator(r.Locator, page)
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
	return nil
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
		if resp.Status() != r.Code {
			t.Fatalf("\n\n%s\n assertStatus failed: expected status code %d, got %d for URL %s\n\n", r.ToErrorLine(), r.Code, resp.Status(), resp.URL())
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

func (r *AssertPath) Execute(t *testing.T, ctx map[string]any, page playwright.Page, resps []playwright.Response) error {
	parsedURL, err := url.Parse(page.URL())
	if err != nil {
		t.Fatalf("\n\n%s\n failed to parse page URL: %v\n\n", r.ToErrorLine(), err)
	}
	if parsedURL.Path != r.URLPath {
		t.Fatalf("\n\n%s\n assertPath failed: expected path %s, got %s\n\n", r.ToErrorLine(), r.URLPath, parsedURL.Path)
	}
	return nil
}

func (r *AssertPath) ToErrorLine() string {
	return fmt.Sprintf("%s:%d: expected path to be %s", r.TestPath, r.TestLine, r.URLPath)
}
